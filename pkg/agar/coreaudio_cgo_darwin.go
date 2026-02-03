//go:build darwin

/*
   Copyright Mycophonic.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package agar

/*
#cgo LDFLAGS: -framework AudioToolbox -framework CoreFoundation

#include <AudioToolbox/AudioToolbox.h>
#include <stdlib.h>
#include <string.h>

// mem_reader provides AudioFileOpenWithCallbacks access to in-memory data.
typedef struct {
	const char *data;
	int64_t     size;
	int64_t     pos;
} mem_reader;

static OSStatus mem_read_proc(void *inClientData, SInt64 inPosition, UInt32 requestCount,
                              void *buffer, UInt32 *actualCount) {
	mem_reader *r = (mem_reader *)inClientData;
	if (inPosition >= r->size) {
		*actualCount = 0;
		return noErr;
	}
	int64_t avail = r->size - inPosition;
	UInt32 toRead = (UInt32)(avail < requestCount ? avail : requestCount);
	memcpy(buffer, r->data + inPosition, toRead);
	*actualCount = toRead;
	return noErr;
}

static SInt64 mem_get_size_proc(void *inClientData) {
	return ((mem_reader *)inClientData)->size;
}

// decode_result holds the output of coreaudio_decode.
typedef struct {
	char    *pcm;
	int64_t  pcm_size;
	char     error[256];
} decode_result;

// coreaudio_decode decodes audio data from memory using AudioToolbox.
// AudioFileOpenWithCallbacks auto-detects the container format (M4A, CAF, MP3, FLAC, etc.).
// Returns 0 on success, non-zero on error (message in result->error).
static int coreaudio_decode(const char *input, int64_t input_size, decode_result *result) {
	memset(result, 0, sizeof(*result));

	mem_reader reader = { .data = input, .size = input_size, .pos = 0 };

	AudioFileID audioFile = NULL;
	OSStatus status = AudioFileOpenWithCallbacks(
		&reader, mem_read_proc, NULL, mem_get_size_proc, NULL,
		0, &audioFile);
	if (status != noErr) {
		snprintf(result->error, sizeof(result->error),
			"AudioFileOpenWithCallbacks failed (OSStatus %d)", (int)status);
		return 1;
	}

	ExtAudioFileRef extFile = NULL;
	status = ExtAudioFileWrapAudioFileID(audioFile, false, &extFile);
	if (status != noErr) {
		snprintf(result->error, sizeof(result->error),
			"ExtAudioFileWrapAudioFileID failed (OSStatus %d)", (int)status);
		AudioFileClose(audioFile);
		return 1;
	}

	// Query source format.
	AudioStreamBasicDescription srcFormat;
	UInt32 propSize = sizeof(srcFormat);
	status = ExtAudioFileGetProperty(extFile, kExtAudioFileProperty_FileDataFormat, &propSize, &srcFormat);
	if (status != noErr) {
		snprintf(result->error, sizeof(result->error),
			"cannot read source format (OSStatus %d)", (int)status);
		ExtAudioFileDispose(extFile);
		AudioFileClose(audioFile);
		return 1;
	}

	// Determine output bit depth.
	UInt32 outBits = srcFormat.mBitsPerChannel;
	if (outBits == 0 && srcFormat.mFormatID == kAudioFormatAppleLossless) {
		switch (srcFormat.mFormatFlags) {
		case kAppleLosslessFormatFlag_16BitSourceData: outBits = 16; break;
		case kAppleLosslessFormatFlag_20BitSourceData: outBits = 20; break;
		case kAppleLosslessFormatFlag_24BitSourceData: outBits = 24; break;
		case kAppleLosslessFormatFlag_32BitSourceData: outBits = 32; break;
		default: outBits = 16; break;
		}
	}
	if (outBits == 0) outBits = 16;

	// 20-bit source is output as 24-bit.
	UInt32 clientBits = outBits;
	if (clientBits == 20) clientBits = 24;
	UInt32 bytesPerSample = clientBits / 8;

	// Set client format: interleaved signed LE PCM.
	AudioStreamBasicDescription clientFormat;
	memset(&clientFormat, 0, sizeof(clientFormat));
	clientFormat.mSampleRate       = srcFormat.mSampleRate;
	clientFormat.mFormatID         = kAudioFormatLinearPCM;
	clientFormat.mFormatFlags      = kAudioFormatFlagIsSignedInteger | kAudioFormatFlagIsPacked;
	clientFormat.mBitsPerChannel   = clientBits;
	clientFormat.mChannelsPerFrame = srcFormat.mChannelsPerFrame;
	clientFormat.mBytesPerFrame    = bytesPerSample * srcFormat.mChannelsPerFrame;
	clientFormat.mFramesPerPacket  = 1;
	clientFormat.mBytesPerPacket   = clientFormat.mBytesPerFrame;

	status = ExtAudioFileSetProperty(extFile, kExtAudioFileProperty_ClientDataFormat,
		sizeof(clientFormat), &clientFormat);
	if (status != noErr) {
		snprintf(result->error, sizeof(result->error),
			"cannot set client format (OSStatus %d)", (int)status);
		ExtAudioFileDispose(extFile);
		AudioFileClose(audioFile);
		return 1;
	}

	// Total frame count.
	SInt64 totalFrames = 0;
	propSize = sizeof(totalFrames);
	status = ExtAudioFileGetProperty(extFile, kExtAudioFileProperty_FileLengthFrames, &propSize, &totalFrames);
	if (status != noErr || totalFrames <= 0) {
		snprintf(result->error, sizeof(result->error),
			"cannot determine frame count (OSStatus %d, frames %lld)", (int)status, totalFrames);
		ExtAudioFileDispose(extFile);
		AudioFileClose(audioFile);
		return 1;
	}

	// Allocate output buffer.
	int64_t pcmSize = totalFrames * clientFormat.mBytesPerFrame;
	char *pcm = (char *)malloc(pcmSize);
	if (!pcm) {
		snprintf(result->error, sizeof(result->error), "out of memory (%lld bytes)", pcmSize);
		ExtAudioFileDispose(extFile);
		AudioFileClose(audioFile);
		return 1;
	}

	// Decode loop.
	const UInt32 framesPerRead = 4096;
	int64_t offset = 0;
	SInt64 framesDecoded = 0;

	while (framesDecoded < totalFrames) {
		SInt64 remaining = totalFrames - framesDecoded;
		UInt32 frameCount = (remaining < framesPerRead) ? (UInt32)remaining : framesPerRead;

		AudioBufferList bufList;
		bufList.mNumberBuffers = 1;
		bufList.mBuffers[0].mNumberChannels = srcFormat.mChannelsPerFrame;
		bufList.mBuffers[0].mDataByteSize   = frameCount * clientFormat.mBytesPerFrame;
		bufList.mBuffers[0].mData           = pcm + offset;

		status = ExtAudioFileRead(extFile, &frameCount, &bufList);
		if (status != noErr) {
			snprintf(result->error, sizeof(result->error),
				"ExtAudioFileRead failed (OSStatus %d)", (int)status);
			free(pcm);
			ExtAudioFileDispose(extFile);
			AudioFileClose(audioFile);
			return 1;
		}
		if (frameCount == 0) break;

		offset += frameCount * clientFormat.mBytesPerFrame;
		framesDecoded += frameCount;
	}

	ExtAudioFileDispose(extFile);
	AudioFileClose(audioFile);

	result->pcm = pcm;
	result->pcm_size = offset;
	return 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// CoreAudioDecode decodes audio from memory using macOS AudioToolbox (CGO).
// AudioToolbox auto-detects the container format, supporting ALAC, AAC, MP3, FLAC, and others.
// Returns raw interleaved signed integer little-endian PCM bytes.
func CoreAudioDecode(data []byte) ([]byte, error) {
	var result C.decode_result

	rc := C.coreaudio_decode(
		(*C.char)(unsafe.Pointer(&data[0])), //nolint:gosec // G103: CGO requires unsafe pointer for C interop.
		C.int64_t(len(data)),
		&result,
	)

	if rc != 0 {
		return nil, fmt.Errorf("coreaudio: %s", C.GoString(&result.error[0]))
	}

	pcm := C.GoBytes(unsafe.Pointer(result.pcm), C.int(result.pcm_size)) //nolint:gosec // G103: CGO requires unsafe for C-allocated memory.
	C.free(unsafe.Pointer(result.pcm))                                    //nolint:gosec // G103: freeing C-allocated memory.

	return pcm, nil
}
