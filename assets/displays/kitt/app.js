window.onload = () => {
    const urlParams = new URLSearchParams(window.location.search);
    const audioUrl = urlParams.get('audioURL');

    if (!audioUrl) {
        console.error("audioURL parameter is missing!");
        document.body.innerHTML = '<p style="color:red; font-family: sans-serif;">Error: audioURL not provided.</p>';
        return;
    }
    visualize(audioUrl);
};

function visualize(audioUrl) {
    const audioContext = new (window.AudioContext || window.webkitAudioContext)();
    const analyser = audioContext.createAnalyser();
    
    analyser.fftSize = 32;
    const bufferLength = analyser.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);
    const voiceboxBars = document.querySelectorAll('.voicebox-bar');

    fetch(audioUrl)
        .then(response => response.arrayBuffer())
        .then(arrayBuffer => audioContext.decodeAudioData(arrayBuffer))
        .then(decodedAudio => {
            const source = audioContext.createBufferSource();
            source.buffer = decodedAudio;
            source.connect(analyser);
            analyser.connect(audioContext.destination);
            source.start(0);
            source.onended = () => {
                // Optional: close the window when the audio finishes.
                // This might require integration with your desktop framework (e.g., Wails, Electron).
                // window.close(); 
            };
            draw();
        })
        .catch(err => {
            console.error('Error processing audio:', err);
            document.body.innerHTML = `<p style="color:red; font-family: sans-serif;">Error processing audio: ${err.message}</p>`;
        });

    function draw() {
        if (audioContext.state === 'closed') return;
        requestAnimationFrame(draw);
        analyser.getByteTimeDomainData(dataArray);
        let sum = 0;
        for (let i = 0; i < bufferLength; i++) {
            sum += Math.abs(dataArray[i] - 128); // 128 is the zero-point
        }
        const average = sum / bufferLength;
        const scale = Math.min(average / 40, 1.0); // Tweak '40' for sensitivity
        voiceboxBars.forEach(bar => {
            const randomScale = scale * (0.8 + Math.random() * 0.2);
            bar.style.transform = `scaleY(${randomScale})`;
        });
    }
}