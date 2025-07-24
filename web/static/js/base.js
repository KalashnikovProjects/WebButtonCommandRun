const WebsocketSendInterval = 50

let scrollSumDelta = 0
let scrollStartLeft = 0
let scrollTargetLeft = 0

let consoleUsing
const apiBase = "/api/v1/"
let commandsList
let commandId = -1
let currentCommand
let terminalWebsocket


function saveFile(name, blob) {
    const url = URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = url;
    a.download = name;
    document.body.appendChild(a);
    a.click();

    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

function smoothHorizontalScroll(e) {
    e.preventDefault();
    e.stopPropagation();
    let el = e.currentTarget;
    if (el.scrollLeft === scrollTargetLeft) scrollSumDelta = 0
    if (scrollSumDelta === 0) scrollStartLeft = el.scrollLeft
    scrollSumDelta -= e.deltaY
    const maxScrollLeft = el.scrollWidth - el.clientWidth;
    const left = Math.min(maxScrollLeft, Math.max(0, scrollStartLeft - scrollSumDelta));
    scrollTargetLeft = left
    el.scrollTo({left, behavior: "smooth"})
}

function toggleMenu(event) {
    let elem = document.getElementById("right-menu")
    if (elem.classList.contains("hidden")) {
        elem.classList.remove("hidden")
    } else {
        elem.classList.add("hidden")
    }
}

function runCommand(event) {
    fitAddon.fit()
    term.resize(term.cols - 2, term.rows)
    terminalWebsocket = new WebSocket(`${apiBase}ws/commands/${commandId}`)
    document.getElementById("command-up-terminal").innerText = currentCommand.name
    let interval
    terminalWebsocket.onopen = (event) => {
        term.write('\x1b[?25h')
        term.reset()
        term.writeln("> " + currentCommand.command)
        commandRunning = true
        term.options.disableStdin = false
        document.body.classList.add("terminal-opened")
        terminalWebsocket.send(JSON.stringify({"message-type": "options", "options": {
                "rows": term.rows,
                "cols": term.cols,
            }}));
        interval = setInterval(() => {
            if (commandRunning && termInputedText && termInputedText.length !== 0) {
                terminalWebsocket.send(JSON.stringify({"message-type": "terminal-input", "data": termInputedText.join("")}));
                termInputedText = [];
            }
        }, WebsocketSendInterval)

    }
    terminalWebsocket.onmessage = (event) => {
        try {
            let data = JSON.parse(event.data)
            switch (data["message-type"]) {
                case "data":
                    console.log("get data", data)
                    term.write(data.data);
                    break
            }
        } catch (e) {}
    }
    terminalWebsocket.onclose = (event) => {
        if (interval) {
            clearInterval(interval)
        }
        term.write('\x1b[?25l')
        termInputedText = []
        commandRunning = false
        term.options.disableStdin = true
        try {
            switch (event.code) {
                case 1000:
                    term.writeln('\n');
                    term.writeln(`\x1b[1;32mFinished\x1b[0m`)
                    break
                case 4001:
                    return
                default:
                    term.writeln('\n');
                    term.writeln(`\x1b[1;31mDisconected: ${event.code} ${event.reason}\x1b[0m`)
                }
        } catch (e) {}
    }
}

function closeTerminal(event) {
    terminalWebsocket.close(4001, "terminal closed from frontend")
    document.body.classList.remove("terminal-opened")
}

function restartCommand(event) {
    terminalWebsocket.close(4001, "terminal closed from frontend")
    runCommand(event)
}

function saveConfig(event) {
    fetch(`${apiBase}json-config`, {
        method: "GET"
    }).then(
        response => response.json()
    ).then(
        data => {
            const jsonString = JSON.stringify(data, null, 2);
            saveFile("commands-config.json", new Blob([jsonString], { type: 'application/json' }))
        }
    )
}

function importConfig(event) {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'application/json';
    input.style.display = 'none';

    input.addEventListener('change', async () => {
        const file = input.files[0];
        if (!file) return;

        try {
            const text = await file.text();
            const json = JSON.parse(text);
            if (json["using-console"] !== consoleUsing) {
                const popup = document.createElement('div');
                popup.id = 'popup';
                popup.innerHTML = `
                  <div class="popup-backdrop"></div>
                  <div class="popup-content">
                    <p>Config use another console: <b>${json["using-console"]}</b> (need: <b>${consoleUsing}</b>). Do import?</p>
                    <div class="popup-buttons">
                      <button id="popup-cancel-btn" class="normal-button red-button">Cancel</button>
                      <button id="popup-confirm-btn" class="normal-button">Import</button>
                    </div>
                  </div>
                `;
                document.body.appendChild(popup);
                document.getElementById('popup-confirm-btn').onclick = function() {
                  fetch(`${apiBase}json-config`, {
                      method: "POST",
                      headers: {
                          'Content-Type': 'application/json'
                      },
                      body: JSON.stringify(json)
                  }).then(
                      response => {
                          commandId = 0
                          let prom = loadCommands();
                          prom.then(renderCommandsList)
                          prom.then(loadCommand)
                      }
                  ).catch(err => {
                      console.error('Ошибка:', err);
                      alert('Ошибка: ' + err.message);
                  });
                  document.body.removeChild(popup);
                };
                document.getElementById('popup-cancel-btn').onclick = function() {
                  document.body.removeChild(popup);
                };
                return;
            }
            fetch(`${apiBase}json-config`, {
                method: "POST",
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(json)
            }).then(
                response => {
                    commandId = 0
                    let prom = loadCommands();
                    prom.then(renderCommandsList)
                    prom.then(loadCommand)
                }
            ).catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка: ' + err.message);
            })
        } catch (err) {
            console.error('Ошибка:', err);
            alert('Ошибка: ' + err.message);
        }
    });

    document.body.appendChild(input);
    input.click();
    document.body.removeChild(input);
}

function renderRunButtonContainer() {
    document.querySelector(".big-run-button-container").innerHTML = `
        <button id="big-run-button" class="round-button">
            <img src="../static/vectors/arrow.svg" alt="Run command"/>
        </button>
        <div class="command-info">
            <input id="main-name-input" type="text" class="command-text command-name" placeholder="Command Name" spellcheck="false">
            <button id="edit-button" class="normal-button small-button">Edit</button>
            <br/>
            <textarea id="main-command-input" class="command-text main-command-text" placeholder="command" spellcheck="false" onWheel="smoothHorizontalScroll(event)"></textarea>
        </div>`

    document.getElementById("edit-button").addEventListener("click", editCommand)
    document.getElementById("big-run-button").addEventListener("click", runCommand)
    document.getElementById("main-name-input").value = currentCommand.name
    document.getElementById("main-command-input").value = currentCommand.command

    document.getElementById("main-name-input").addEventListener("blur", (event) => {
        if (currentCommand.name !== event.target.value) {
            fetch(`${apiBase}commands/${commandId}`, {
                method: "PATCH",
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    name: String(event.target.value),
                })
            }).then(response => {
                currentCommand.name = event.target.value
                commandsList[commandId] = currentCommand
                renderCommandsList()
            }).catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка: ' + err.message);
            });
        }
    })
    document.getElementById("main-command-input").addEventListener("blur", (event) => {
        if (currentCommand.command !== event.target.value) {
            fetch(`${apiBase}commands/${commandId}`, {
                method: "PATCH",
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    command: String(event.target.value),
                })
            }).then(response => {
                currentCommand.command = event.target.value
            }).catch(err => {
                console.error('Ошибка:', err);
                alert('Ошибка: ' + err.message);
            });
        }
    })
}

function renderNoCommands() {
    document.querySelector(".big-run-button-container").innerHTML = `
        <h2>No created commands</h2>
        <h3 style="margin-top: 5px">Create first to start</h3>
<!--        <img height="20%" src="../static/vectors/down-arrow.svg" alt="Down arrow"/>-->
        <button id="add-new-command-center" class="normal-button" style="height: auto;padding: 12px 25px;">
            Add command
        </button>
    `
    document.getElementById("add-new-command-center").addEventListener("click", addNewCommand)
}

function initPage() {
    fetch(`${apiBase}console-using`, {
        method: "GET"
    }).then(response => response.text()).then(text => {
        consoleUsing = text
        switch (consoleUsing) {
            case "cmd":
                document.getElementById("using-console-indicator").innerHTML = "using cmd <img src=\"../static/vectors/console-cmd.svg\" alt=\"\"/>"
                break
            case "sh":
                document.getElementById("using-console-indicator").innerHTML = "using cmd <img src=\"../static/vectors/console-bash.svg\" alt=\"\"/>"
                break
            default:
                document.getElementById("using-console-indicator").innerHTML = "unknown console"
        }
    }).catch(err => {
        console.error('Ошибка:', err);
        alert('Ошибка: ' + err.message);
    })

    document.getElementById("open-menu-button").addEventListener("click", toggleMenu)
    document.getElementById("close-button").addEventListener("click", closeTerminal)
    document.getElementById("restart-button").addEventListener("click", restartCommand)
    document.getElementById("save-config-button").addEventListener("click", saveConfig)
    document.getElementById("import-config-button").addEventListener("click", importConfig)
    document.getElementById("command-list").addEventListener("wheel", smoothHorizontalScroll, { passive: false });

    let a = window.location.hash.split("-")

    if (a.length === 2) {
        commandId = Number(a[1])
    } else {
        commandId = 0
    }
    loadCommands().then(() => renderCommandsList(true))
    loadCommand()
}

function editCommand(event) {
    const popup = document.createElement('div');
    popup.id = 'popup';
    popup.innerHTML = `
                  <div class="popup-backdrop hidden"></div>
                  <div class="popup-content big-popup hidden">
                    <h2>Edit command</h2>
                       <div class="input-line">
                           <button id="delete-command-btn" class="normal-button red-button">Delete <img src="../static/vectors/delete.svg" alt=""/></button>
                       </div>
                        <div class="popup-buttons" style="margin-top: 30px">
                          <button id="popup-confirm-btn" class="normal-button">Close</button>
                    </div>
                  </div>`;
    document.body.appendChild(popup);

    setTimeout(() => {
        document.querySelector(".popup-backdrop").classList.remove("hidden")
        document.querySelector(".popup-content").classList.remove("hidden")
    }, 20)
    document.getElementById("delete-command-btn").addEventListener("click", () => {
        fetch(`${apiBase}commands/${commandId}`, {
            method: "DELETE"
        }).then(response => {
            commandsList.splice(commandId, 1);
            commandId = 0
            loadCommand()
            renderCommandsList()
            document.querySelector(".popup-backdrop").classList.add("hidden")
            document.querySelector(".popup-content").classList.add("hidden")
            setTimeout(
                () => {
                    document.body.removeChild(popup)
                },
                300
            )
        }).catch(err => {
            console.error('Ошибка:', err);
            alert('Ошибка: ' + err.message);
        })
    })

    document.getElementById('popup-confirm-btn').onclick = function() {
        document.querySelector(".popup-backdrop").classList.add("hidden")
        document.querySelector(".popup-content").classList.add("hidden")
        setTimeout(
            () => {
                document.body.removeChild(popup)
            },
            300
        )
    };
}

function addNewCommand(event) {
    const popup = document.createElement('div');
    popup.id = 'popup';
    popup.innerHTML = `
                  <div class="popup-backdrop hidden"></div>
                  <div class="popup-content big-popup hidden">
                    <h2>Add new command</h2>
                    <div class="input-line">
                            <label for="popup-input-name">Name</label>
                            <input id="popup-input-name" required type="text" class="command-text command-name" spellcheck="false" oninput="this.size = this.value.length" \>
                        </div>
                        <div class="input-line">
                            <label for="popup-input-name">Command</label>
                            <input id="popup-input-command" required type="text" class="command-text" spellcheck="false" oninput="this.size = this.value.length" \>
                        </div>
                        <div class="popup-buttons" style="margin-top: 30px">
                          <button id="popup-cancel-btn" class="normal-button red-button">Cancel</button>
                          <button id="popup-confirm-btn" class="normal-button">Add</button>
                    </div>
                  </div>`;
    document.body.appendChild(popup);
    setTimeout(() => {
        document.querySelector(".popup-backdrop").classList.remove("hidden")
        document.querySelector(".popup-content").classList.remove("hidden")
    }, 20)
    document.getElementById('popup-confirm-btn').onclick = function() {
        let name = document.getElementById("popup-input-name").value
        let command = document.getElementById("popup-input-command").value
        fetch(`${apiBase}commands/`, {
            method: "POST",
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                name: String(name),
                command: String(command)
            })
        }).then(
            response => {loadCommands().then(renderCommandsList).then(() => selectCommand(commandsList.length - 1))}
        ).catch(err => {
            console.error('Ошибка:', err);
            alert('Ошибка: ' + err.message);
        });
        document.querySelector(".popup-backdrop").classList.add("hidden")
        document.querySelector(".popup-content").classList.add("hidden")
        setTimeout(
            () => {
                document.body.removeChild(popup)
            },
            300
        )
    };
    document.getElementById('popup-cancel-btn').onclick = function() {
        document.querySelector(".popup-backdrop").classList.add("hidden")
        document.querySelector(".popup-content").classList.add("hidden")
        setTimeout(
            () => {
                document.body.removeChild(popup)
            },
            300
        )
    };
}

function selectButtonIcons(id) {
    for (let button of document.querySelectorAll(".command-list li:not(#new-command-btn)")) {
        if (Number(button.id.split("-")[2]) === id) {
            button.querySelector("img").src = "../static/vectors/selected.svg"
            button.classList.add("selected")
        } else {
            button.querySelector("img").src = "../static/vectors/arrow.svg"
            button.classList.remove("selected")
        }
    }
}

function selectCommand(num) {
    if (document.body.classList.contains("terminal-opened")) {
        closeTerminal()
    }
    commandId=num;
    selectButtonIcons(num);
    loadCommand()
}

function loadCommand() {
    location.hash=`command-${commandId}`;
    fetch(`${apiBase}commands/${commandId}`, {
        method: "GET"
    }).then(response => response.json()).then(data => {
        currentCommand = data
        renderRunButtonContainer()
    }).catch(err => {
        if (!commandsList || commandsList.length === 0) {
            commandId = -1
            location.hash=`no-commands`;
            renderNoCommands()
        }
        commandId = 0
        location.hash=`command-0`;
        fetch(`${apiBase}commands/${commandId}`, {
            method: "GET"
        }).then(response => response.json()).then(data => {
            currentCommand = data
            renderRunButtonContainer()
        }).catch(err => {
            commandId = -1
            location.hash=`no-commands`;
            renderNoCommands()
        })
    })
}

function renderCommandsList(withAnimation = false) {
    if (withAnimation) {
        document.getElementById("command-list").classList.add("gap-0-for-animation")
    }
    let listElem = document.getElementById("command-list")
    listElem.innerHTML = ""
    let num = 0
    for (let i of commandsList) {
        let elem = document.createElement("li")
        elem.id = `command-btn-${num}`
        console.log(i, commandsList)
        if (withAnimation) {
            elem.classList.add("width-0-for-animation")
        }
        elem.innerHTML = `<button class="round-button" onclick="selectCommand(${num})">
                    <img width="80%" src="../static/vectors/arrow.svg" alt="Select command"/>
                </button>
                <p class="command-text command-name" onWheel="smoothHorizontalScroll(event)">
                    ${i.name}
                </p>`
        listElem.appendChild(elem)
        num += 1
    }
    let elem = document.createElement("li")
    elem.id = `new-command-btn`
    if (withAnimation) {
        elem.classList.add("width-0-for-animation")
    }
    elem.innerHTML = `<button class="round-button">
                    <img width="100%" src="../static/vectors/plus.svg" alt="Add new command"/>
                </button>
                <p class="command-text command-name" onWheel="smoothHorizontalScroll(event)">
                    Add new
                </p>`
    elem.addEventListener("click", addNewCommand)
    listElem.appendChild(elem)
    selectButtonIcons(commandId)
    if (withAnimation) {
        setTimeout(() => {
            document.getElementById("command-list").classList.remove("gap-0-for-animation")
            for (elem of document.querySelectorAll(".width-0-for-animation")) {
                elem.classList.remove("width-0-for-animation")
            }
        }, 30)
    }
}

async function loadCommands() {
    try {
        let response = await fetch(`${apiBase}commands/`, {
            method: "GET"
        })
        commandsList = await response.json()
    } catch (err) {
        console.error('Ошибка:', err);
        alert('Ошибка: ' + err.message);
    }
}

initPage()
