let command = '';
let term;

function initTerminal() {
    term = new Terminal({
        theme: {
            background: '#1e1e1e',
            foreground: 'white', // #6b7fd7
            cursor: 'white',
        },
        fontFamily: 'Fira Code, monospace',
        fontSize: 14,
    });
    const fitAddon = new FitAddon.FitAddon();
    term.loadAddon(fitAddon);
    term.open(document.getElementById('terminal'));
    fitAddon.fit();

    term.writeln('Text will be here soon');
    term.write('$ ');
    term.onKey(onTermKey)
}


function onTermKey(e){
    const char = e.key;

    if (e.domEvent.key === 'Enter') {
        term.writeln('');
        handleCommand(command);
        command = '';
        term.write('$ ');

    } else if (e.domEvent.key === 'Backspace') {
        if (command.length > 0) {
            command = command.slice(0, -1);
            term.write('\b \b');
        }

    } else if (e.domEvent.key.length === 1) {
        command += char;
        term.write(char);
    }
}

function handleCommand(cmd) {
}


initTerminal()