let termInputedText = [];
let term;
let commandRunning = false
const fitAddon = new FitAddon.FitAddon();

function initTerminal() {
    let termElem = document.getElementById('terminal')
    term = new Terminal({
        theme: {
            background: '#1e1e1e',
            foreground: 'white', // #6b7fd7
            cursor: 'white',
        },
        fontFamily: 'Fira Code, monospace',
        fontSize: 18,
        cursorBlink: true,
        cursorStyle: 'block',
        allowTransparency: true,
        cols: 80,
        rows: 24,
        convertEol: true,
        disableStdin: true, // Initially disabled until command starts
    });
    term.loadAddon(fitAddon);
    term.open(termElem);
    fitAddon.fit()
    term.resize(term.cols - 2, term.rows)

    term.onData(onTermData)
    // term.onKey(onKey)
}

function onTermData(data){
    // if (data === "\x03") {
    //     // CTRL + C
    //     console.log(term.getSelection())
    //     // Why term.getSelection() always return empty sting? Dont work
    //     navigator.clipboard.writeText(term.getSelection())
    // }
    if (commandRunning) {
        if (data === "\x16") {
            // CTRL + V
            navigator.clipboard.readText().then(clipText => {termInputedText.push(clipText)})
        }
        termInputedText.push(data);
    }
}


// function onKey(e){
//     let char = e.key
//     if (!commandRunning) {
//         return
//     }
//     if (e.domEvent.key === 'Enter') {
//         term.writeln("");
//     } else if (e.domEvent.key === 'Backspace') {
//         term.write('\b \b');
//     } else {
//         term.write(char);
//     }
// }

initTerminal()