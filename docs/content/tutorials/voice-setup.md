---
title: "Voice Setup Tutorial"
description: "Complete guide to download and configure voices for Focus Helper"
date: 2024-09-09
weight: 20
---

# Voice Setup Tutorial

This tutorial will guide you through downloading and configuring voices for the Focus Helper application.

## Overview

Focus Helper uses text-to-speech (TTS) voices to provide audio feedback and notifications. The application supports multiple voice formats and languages.

## Supported Voice Formats

- **ONNX Runtime**: `.onnx` files with `.onnx.json` metadata
- **eSpeak**: Built-in system voices
- **Festival**: Linux system voices
- **Custom**: Any TTS engine that can be integrated

## Step 1: Download Voices

### Option 1: Using Pre-trained ONNX Models

1. **Visit the Coqui TTS Model Hub**:
   - Go to [https://huggingface.co/coqui-ai/TTS-models](https://huggingface.co/coqui-ai/TTS-models)
   - Browse available models

2. **Download a Model**:
   ```bash
   # Example: Download a Portuguese model
   wget https://huggingface.co/coqui-ai/TTS-models/resolve/main/tts_models/multilingual/multi-dataset/xtts_v2/model.pth
   ```

3. **Convert to ONNX** (if needed):
   ```python
   from TTS.api import TTS
   import torch
   
   # Load the model
   tts = TTS("tts_models/multilingual/multi-dataset/xtts_v2")
   
   # Convert to ONNX
   tts.export_onnx("path/to/output/model.onnx")
   ```

### Option 2: Using Coqui TTS CLI

1. **Install Coqui TTS**:
   ```bash
   pip install TTS
   ```

2. **Download and Convert**:
   ```bash
   # Download a specific model
   tts --model_name "tts_models/pt/cv/vits" --text "Hello world" --out_path "output.wav"
   
   # List available models
   tts --list_models
   ```

### Option 3: Using Pre-converted ONNX Models

1. **Download from Community Sources**:
   - Check the [Focus Helper Community Voices](https://github.com/robertocpaes/focus-helper-voices) repository
   - Download pre-converted `.onnx` and `.onnx.json` files

2. **Place in Voices Directory**:
   ```bash
   mkdir -p voices
   cp downloaded_model.onnx voices/
   cp downloaded_model.onnx.json voices/
   ```

## Step 2: Configure Voice in Focus Helper

### Method 1: Using Profiles Configuration

1. **Edit your profile** in `profiles.json`:
   ```json
   {
     "name": "default",
     "voice": {
       "enabled": true,
       "engine": "onnx",
       "model_path": "./voices/pt_BR-cadu-medium.onnx",
       "model_config": "./voices/pt_BR-cadu-medium.onnx.json",
       "speaker_id": "default",
       "language": "pt-BR"
     }
   }
   ```

2. **Restart Focus Helper**:
   ```bash
   ./focushelper -profile default
   ```

### Method 2: Using Environment Variables

1. **Set voice configuration**:
   ```bash
   export FOCUSHELPER_VOICE_ENABLED=true
   export FOCUSHELPER_VOICE_ENGINE=onnx
   export FOCUSHELPER_VOICE_MODEL_PATH=./voices/pt_BR-cadu-medium.onnx
   export FOCUSHELPER_VOICE_MODEL_CONFIG=./voices/pt_BR-cadu-medium.onnx.json
   ```

2. **Run Focus Helper**:
   ```bash
   ./focushelper
   ```

## Step 3: Test Your Voice Configuration

### Test Voice Output

1. **Run Focus Helper in debug mode**:
   ```bash
   ./focushelper -debug
   ```

2. **Trigger a voice command**:
   - Say your activation phrase (default: "torre controle comando")
   - Say "tempo" to test time announcement
   - Say "check" to test status announcement

3. **Check logs** for voice-related messages:
   ```bash
   tail -f focus_helper.log | grep -i voice
   ```

### Test Different Voices

1. **Switch between voices** by changing the model path in your profile
2. **Test voice quality** with different text samples
3. **Adjust voice parameters** if supported by your model

## Step 4: Advanced Configuration

### Voice Parameters

Configure voice parameters in your profile:

```json
{
  "voice": {
    "enabled": true,
    "engine": "onnx",
    "model_path": "./voices/pt_BR-cadu-medium.onnx",
    "model_config": "./voices/pt_BR-cadu-medium.onnx.json",
    "speaker_id": "default",
    "language": "pt-BR",
    "speed": 1.0,
    "pitch": 1.0,
    "volume": 0.8,
    "sample_rate": 22050,
    "chunk_size": 1024
  }
}
```

### Multiple Voice Support

1. **Create multiple profiles** with different voices:
   ```json
   [
     {
       "name": "portuguese",
       "voice": {
         "model_path": "./voices/pt_BR-cadu-medium.onnx"
       }
     },
     {
       "name": "english",
       "voice": {
         "model_path": "./voices/en_US-male-medium.onnx"
       }
     }
   ]
   ```

2. **Switch profiles**:
   ```bash
   ./focushelper -profile portuguese
   ./focushelper -profile english
   ```

## Troubleshooting

### Common Issues

1. **Voice not working**:
   - Check if the model file exists
   - Verify file permissions
   - Check logs for error messages

2. **Poor voice quality**:
   - Try a different model
   - Adjust voice parameters
   - Check audio system configuration

3. **Model loading errors**:
   - Verify ONNX model compatibility
   - Check model configuration file
   - Ensure all dependencies are installed

### Debug Commands

```bash
# Check voice configuration
./focushelper -debug -voice-test

# Test specific voice file
./focushelper -voice-test -model ./voices/your_model.onnx

# List available voices
./focushelper -list-voices

# Check voice engine status
./focushelper -voice-status
```

## Voice Model Recommendations

### Portuguese (Brazil)
- **Coqui XTTS v2**: High quality, multilingual
- **Tacotron2 + WaveGlow**: Good for Portuguese
- **FastSpeech2**: Fast synthesis

### English
- **Coqui XTTS v2**: Multilingual with English support
- **Tacotron2**: High quality English
- **FastSpeech2**: Fast and efficient

### Other Languages
- **Coqui XTTS v2**: Supports 13+ languages
- **Multilingual models**: Check Coqui TTS documentation

## Community Resources

- [Focus Helper Community Voices](https://github.com/robertocpaes/focus-helper-voices)
- [Coqui TTS Documentation](https://tts.readthedocs.io/)
- [ONNX Runtime Documentation](https://onnxruntime.ai/)
- [Voice Model Hub](https://huggingface.co/coqui-ai/TTS-models)

## Next Steps

After setting up voices, you can:

1. **Customize voice responses** in the language files
2. **Add new voice commands** in the voice configuration
3. **Integrate with external TTS services** via webhooks
4. **Create custom voice models** for specific use cases

For more advanced configuration, see the [Advanced Configuration Guide](../configuration/advanced.md).
