const WebsocketSendInterval = 50

let consoleUsing
const apiBase = "/api/v1/"
let commandsList
let commandId = -1
let currentCommand
let terminalWebsocket

initPage();

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
    const el = e.currentTarget;
    const ds = el.dataset;
    let sumDelta = ds.scrollSumDelta ? Number(ds.scrollSumDelta) : 0;
    let startLeft = ds.scrollStartLeft ? Number(ds.scrollStartLeft) : 0;
    let targetLeft = ds.scrollTargetLeft ? Number(ds.scrollTargetLeft) : el.scrollLeft;
    if (el.scrollLeft === targetLeft) sumDelta = 0;
    if (sumDelta === 0) startLeft = el.scrollLeft;
    sumDelta -= e.deltaY;
    const maxScrollLeft = el.scrollWidth - el.clientWidth;
    const left = Math.min(maxScrollLeft, Math.max(0, startLeft - sumDelta));
    ds.scrollSumDelta = String(sumDelta);
    ds.scrollStartLeft = String(startLeft);
    ds.scrollTargetLeft = String(left);
    el.scrollTo({ left, behavior: "smooth" });
}

function toggleMenu(event) {
    const elem = document.getElementById("right-menu");
    elem.classList.toggle("hidden");
}

function getWebSocketProtocol() {
    return location.protocol === "https:" ? "wss" : "ws";
}

function runCommand(event) {
    if (commandId === -1 || !currentCommand) {
        return;
    }
    if (typeof fitAddon !== 'undefined' && typeof term !== 'undefined') {
        try {
            fitAddon.fit();
            term.resize(term.cols - 2, term.rows);
        } catch (_) {}
    }
    const protocol = getWebSocketProtocol();
    terminalWebsocket = new WebSocket(`${protocol}://${location.host}${apiBase}ws/commands/${commandId}`);
    document.getElementById("command-up-terminal").innerText = currentCommand.name;
    let interval;
    terminalWebsocket.onopen = () => {
        term.write('\x1b[?25h');
        term.reset();
        term.writeln("> " + currentCommand.command);
        commandRunning = true;
        term.options.disableStdin = false;
        document.body.classList.add("terminal-opened");
        terminalWebsocket.send(JSON.stringify({
            "message-type": "options",
            "options": { "rows": term.rows, "cols": term.cols }
        }));
        interval = setInterval(() => {
            if (commandRunning && termInputedText && termInputedText.length !== 0) {
                terminalWebsocket.send(JSON.stringify({
                    "message-type": "terminal-input",
                    "data": termInputedText.join("")
                }));
                termInputedText = [];
            }
        }, WebsocketSendInterval);
    };
    terminalWebsocket.onmessage = (event) => {
        try {
            let data = JSON.parse(event.data);
            switch (data["message-type"]) {
                case "data":
                    console.log("get data", data);
                    term.write(data.data);
                    break
            }
        } catch (_) {}
    };
    terminalWebsocket.onclose = (event) => {
        if (interval) {
            clearInterval(interval);
        }
        term.write('\x1b[?25l');
        termInputedText = [];
        commandRunning = false;
        term.options.disableStdin = true;
        try {
            switch (event.code) {
                case 1000:
                    term.writeln('\n');
                    term.writeln(`\x1b[1;32mFinished\x1b[0m`);
                    break
                case 4001:
                    return
                default:
                    term.writeln('\n');
                    term.writeln(`\x1b[1;31mDisconected: ${event.code} ${event.reason}\x1b[0m`);
            }
        } catch (_) {}
    };
}

function closeTerminal(event) {
    terminalWebsocket.close(4001, "terminal closed from frontend");
    document.body.classList.remove("terminal-opened");
}

function restartCommand(event) {
    terminalWebsocket.close(4001, "terminal closed from frontend");
    runCommand(event);
}

function saveConfig(event) {
    fetch(`${apiBase}json-config`, {
        method: "GET"
    }).then(
        async response => {
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Server error: ${response.status} - ${errorText}`);
            }
            return response.json()
        }
    ).then(
        data => {
            const jsonString = JSON.stringify(data, null, 2);
            saveFile("commands-config.json", new Blob([jsonString], { type: 'application/json' }));
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
            if (json.usingConsole && json.usingConsole !== consoleUsing) {
                const popup = document.createElement('div');
                popup.id = 'popup';
                popup.innerHTML = `
                  <div class="popup-backdrop"></div>
                  <div class="popup-content">
                    <p>Config use another console: <b id="using-console-warn"></b> (need: <b id="need-console-warn"></b>). Do import?</p>
                    <div class="popup-buttons">
                      <button id="popup-cancel-btn" class="normal-button red-button">Cancel</button>
                      <button id="popup-confirm-btn" class="normal-button">Import</button>
                    </div>
                  </div>
                `;
                document.body.appendChild(popup);

                document.getElementById("using-console-warn").innerText = json.usingConsole;
                document.getElementById("need-console-warn").innerText = consoleUsing;
                document.getElementById('popup-confirm-btn').onclick = function() {
                  fetch(`${apiBase}json-config`, {
                      method: "POST",
                      headers: {
                          'Content-Type': 'application/json'
                      },
                      body: JSON.stringify(json),
                  }).then(
                      async response => {
                          if (!response.ok) {
                              const errorText = await response.text();
                              throw new Error(`Server error: ${response.status} - ${errorText}`);
                          }
                          let prom = loadCommands().then(() => {
                              if (commandsList.length !== 0) {
                                  commandId = commandsList[0].id;
                              }
                          });
                          prom.then(renderCommandsList);
                          prom.then(loadCommand);
                      }
                  ).catch(err => {
                      console.error('Ошибка:', err);
                      showErrorPopup(
                          'Ошибка импорта конфигурации',
                          'Не удалось импортировать конфигурацию.',
                          err.message
                      );
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
                async response => {
                    if (!response.ok) {
                        const errorText = await response.text();
                        throw new Error(`Server error: ${response.status} - ${errorText}`);
                    }
                    let prom = loadCommands().then(() => {
                        if (commandsList.length !== 0) {
                            commandId = commandsList[0].id;
                        }
                    });
                    prom.then(renderCommandsList);
                    prom.then(loadCommand);
                }
            ).catch(err => {
                console.error('Ошибка:', err);
                showErrorPopup(
                    'Ошибка импорта конфигурации',
                    'Не удалось импортировать конфигурацию.',
                    err.message
                );
            })
        } catch (err) {
            console.error('Ошибка:', err);
            showErrorPopup(
                'Ошибка импорта конфигурации',
                'Не удалось прочитать файл конфигурации. Убедитесь, что файл не поврежден.',
                err.message
            );
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

    document.getElementById("edit-button").addEventListener("click", editCommand);
    document.getElementById("big-run-button").addEventListener("click", runCommand);
    document.getElementById("main-name-input").value = currentCommand.name;
    document.getElementById("main-command-input").value = currentCommand.command;

    document.getElementById("main-name-input").addEventListener("blur", (event) => {
        if (currentCommand.name !== event.target.value) {
            if (event.target.value === "") {
                event.target.value = currentCommand.name;
                showErrorPopup(
                    'Неверный ввод',
                    'Название команды не должно быть пустым',
                );
            }
            fetch(`${apiBase}commands/${commandId}`, {
                method: "PATCH",
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    name: String(event.target.value),
                })
            }).then(async response => {
                if (!response.ok) {
                    const errorText = await response.text();
                    event.target.value = currentCommand.name;
                    throw new Error(`Server error: ${response.status} - ${errorText}`);
                }
                currentCommand.name = String(event.target.value);
                if (currentCommand.name === "") {
                    loadCommand().then(() => {
                        for (let i=0; i<commandsList.length; i++) {
                            if (commandsList[i].id === commandId) {
                                commandsList[i] = currentCommand;
                            }
                        }
                        renderCommandsList();
                    })
                } else {
                    for (let i=0; i<commandsList.length; i++) {
                        if (commandsList[i].id === commandId) {
                            commandsList[i] = currentCommand;
                        }
                    }
                    renderCommandsList();
                }
            }).catch(err => {
                console.error('Ошибка:', err);
                showErrorPopup(
                    'Ошибка обновления команды',
                    'Не удалось обновить название команды.',
                    err.message
                );
            });
        }
    });
    document.getElementById("main-command-input").addEventListener("blur", (event) => {
        if (currentCommand.command !== event.target.value) {
            if (event.target.value === "") {
                event.target.value = currentCommand.command;
                showErrorPopup(
                    'Неверный ввод',
                    'Команда не должна быть пустой',
                );
            }
            fetch(`${apiBase}commands/${commandId}`, {
                method: "PATCH",
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    command: String(event.target.value),
                })
            }).then(async response => {
                if (!response.ok) {
                    const errorText = await response.text();
                    event.target.value = currentCommand.command;
                    throw new Error(`Server error: ${response.status} - ${errorText}`);
                }
                currentCommand.command = String(event.target.value);
            }).catch(err => {
                console.error('Ошибка:', err);
                showErrorPopup(
                    'Ошибка обновления команды',
                    'Не удалось обновить команду.',
                    err.message
                );
            });
        }
    });
}

function renderNoCommands() {
    document.querySelector(".big-run-button-container").innerHTML = `
        <h2>No created commands</h2>
        <h3 style="margin-top: 5px">Create first to start</h3>
<!--        <img height="20%" src="../static/vectors/down-arrow.svg" alt="Down arrow"/>-->
        <button id="add-new-command-center" class="normal-button" style="height: auto;padding: 12px 25px;">
            Add command
        </button>
    `;
    document.getElementById("add-new-command-center").addEventListener("click", addNewCommand);
}

function initPage() {
    fetch(`${apiBase}console-using`, {
        method: "GET"
    }).then(async response => {
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
        return response.text()
    }).then(text => {
        consoleUsing = text;
        switch (consoleUsing) {
            case "cmd":
                document.getElementById("using-console-indicator").innerHTML = "using cmd <img src=\"../static/vectors/console-cmd.svg\" alt=\"\"/>";
                break
            case "sh":
                document.getElementById("using-console-indicator").innerHTML = "using sh <img src=\"../static/vectors/console-bash.svg\" alt=\"\"/>";
                break
            default:
                document.getElementById("using-console-indicator").innerHTML = "unknown console";
        }
    }).catch(err => {
        console.error('Ошибка:', err);
        showErrorPopup(
            'Ошибка загрузки консоли',
            'Не удалось определить тип используемой консоли.',
            err.message
        );
    })

    document.getElementById("open-menu-button").addEventListener("click", toggleMenu);
    document.getElementById("close-button").addEventListener("click", closeTerminal);
    document.getElementById("restart-button").addEventListener("click", restartCommand);
    document.getElementById("save-config-button").addEventListener("click", saveConfig);
    document.getElementById("import-config-button").addEventListener("click", importConfig);
    document.getElementById("export-files-button").addEventListener("click", exportFiles);
    document.getElementById("import-files-button").addEventListener("click", importFiles);

    document.getElementById("command-list").addEventListener("wheel", smoothHorizontalScroll, { passive: false });

    let hash = window.location.hash.split("-");

    let prom = loadCommands().then(() => {
        if (hash.length === 2) {
            commandId = parseInt(hash[1]);
            if (isNaN(commandId)) {
                if (commandsList.length !== 0) {
                    commandId = commandsList[0].id;
                }
            }
        } else {
            if (commandsList.length !== 0) {
                commandId = commandsList[0].id;
            }
        }
    });
    prom.then(() => renderCommandsList(true));
    prom.then(loadCommand);
}

function editCommand(event) {
    const popup = document.createElement('div');
    popup.id = 'popup';
    popup.innerHTML = `
                  <div class="popup-backdrop hidden"></div>
                  <div class="popup-content big-popup hidden">
                      <h2>Edit command</h2>
                      <h3 style="text-align: left; margin-bottom: 8px">Files</h3>
                      <div id="files-list">
                          <p>Loading files...</p>
                      </div>
                      <div style="display: flex; gap: 10px; align-items: center">
                          <input type="file" id="file-input" multiple style="display: none;">
                          <button id="select-files-btn" class="normal-button">Upload files</button>
                          <span id="selected-files-info" style="font-size: 0.9em; opacity: 0.7;"></span>
                      </div>
                      <h3 style="text-align: left; margin-bottom: 5px">Command execution dir</h3>
                      <textarea id="command-execution-dir-input" class="command-text main-command-text" spellcheck="false" onWheel="smoothHorizontalScroll(event)" value="${currentCommand.executionDir}">${currentCommand.executionDir}</textarea>
                      <h3 style="text-align: left; margin-bottom: 5px">Delete command</h3>

                      <button id="delete-command-btn" class="normal-button red-button">Delete <img src="../static/vectors/delete.svg" alt=""/></button>
                      <div class="popup-buttons" style="margin-top: 30px">
                          <button id="popup-confirm-btn" class="normal-button">Close</button>
                      </div>
                  </div>`;
    document.body.appendChild(popup);

    setTimeout(() => {
        document.querySelector(".popup-backdrop").classList.remove("hidden");
        document.querySelector(".popup-content").classList.remove("hidden");
    }, 20)

    loadCommandFiles();

    document.getElementById("select-files-btn").addEventListener("click", () => {
        document.getElementById("file-input").click();
    });

    document.getElementById("file-input").addEventListener("change", handleFileSelection);

    document.getElementById("delete-command-btn").addEventListener("click", () => {
        cleanupFileHandlers();
        fetch(`${apiBase}commands/${commandId}`, {
            method: "DELETE"
        }).then(async response => {
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Server error: ${response.status} - ${errorText}`);
            }
            for (let i=0; i<commandsList.length; i++) {
                if (commandsList[i].id === commandId) {
                    commandsList.splice(i, 1);
                }
            }
            loadCommand();
            if (commandsList.length !== 0) {
                commandId = commandsList[0].id;
            }
            renderCommandsList();
            document.querySelector(".popup-backdrop").classList.add("hidden");
            document.querySelector(".popup-content").classList.add("hidden");
            setTimeout(
                () => document.body.removeChild(popup),
                300
            );
        }).catch(err => {
            console.error('Ошибка:', err);
            showErrorPopup(
                'Ошибка удаления команды',
                'Не удалось удалить команду.',
                err.message
            );
        })
    })
    document.getElementById("command-execution-dir-input").addEventListener("blur", (event) => {
        if (currentCommand.executionDir !== event.target.value) {
            if (event.target.value === "") {
                event.target.value = currentCommand.executionDir;
                showErrorPopup(
                    'Неверный ввод',
                    'Директория не должна быть пустой',
                );
            }
            fetch(`${apiBase}commands/${commandId}`, {
                method: "PATCH",
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    executionDir: String(event.target.value),
                })
            }).then(async response => {
                if (!response.ok) {
                    const errorText = await response.text();
                    event.target.value = currentCommand.executionDir;
                    throw new Error(`Server error: ${response.status} - ${errorText}`);
                }
                currentCommand.executionDir = String(event.target.value);
            }).catch(err => {
                console.error('Ошибка:', err);
                showErrorPopup(
                    'Ошибка обновления команды',
                    'Не удалось обновить команду.',
                    err.message
                );
            });
        }
    });
    document.getElementById('popup-confirm-btn').onclick = function() {
        cleanupFileHandlers();
        document.querySelector(".popup-backdrop").classList.add("hidden");
        document.querySelector(".popup-content").classList.add("hidden");
        setTimeout(
            () => document.body.removeChild(popup),
            300
        );
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
        document.querySelector(".popup-backdrop").classList.remove("hidden");
        document.querySelector(".popup-content").classList.remove("hidden");
    }, 20)
    document.getElementById('popup-confirm-btn').onclick = function() {
        let name = document.getElementById("popup-input-name").value;
        let command = document.getElementById("popup-input-command").value;
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
            async response => {
                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(`Server error: ${response.status} - ${errorText}`);
                }
                loadCommands().then(renderCommandsList).then(() => selectCommand(commandsList.at(-1).id))
            }
        ).catch(err => {
            console.error('Ошибка:', err);
            showErrorPopup(
                'Ошибка создания команды',
                'Не удалось создать новую команду.',
                err.message
            );
        });
        document.querySelector(".popup-backdrop").classList.add("hidden");
        document.querySelector(".popup-content").classList.add("hidden");
        setTimeout(
            () => {
                document.body.removeChild(popup);
            },
            300
        );
    };
    document.getElementById('popup-cancel-btn').onclick = function() {
        document.querySelector(".popup-backdrop").classList.add("hidden");
        document.querySelector(".popup-content").classList.add("hidden");
        setTimeout(
            () => {
                document.body.removeChild(popup);
            },
            300
        );
    };
}

function selectButtonIcons(id) {
    for (let button of document.querySelectorAll(".command-list li:not(#new-command-btn)")) {
        if (Number(button.id.split("-")[2]) === id) {
            button.querySelector("img").src = "../static/vectors/selected.svg";
            button.classList.add("selected");
        } else {
            button.querySelector("img").src = "../static/vectors/arrow.svg";
            button.classList.remove("selected");
        }
    }
}

function selectCommand(num) {
    if (document.body.classList.contains("terminal-opened")) {
        closeTerminal();
    }
    commandId=num;
    selectButtonIcons(num);
    loadCommand();
}

function loadCommand() {
    if (commandsList.length === 0) {
        renderNoCommands()
        return
    }
    location.hash=`command-${commandId}`;
    return fetch(`${apiBase}commands/${commandId}`, {
        method: "GET"
    }).then(async response => {
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
        return response.json()
    }).then(data => {
        currentCommand = data;
        renderRunButtonContainer();
    }).catch(err => {
        if (commandsList.length === 0 || commandsList.length === 0) {
            commandId = -1;
            location.hash=`no-commands`;
            renderNoCommands();
        }
        commandId = commandsList[0].id;
        location.hash=`command-${commandId}`;
        fetch(`${apiBase}commands/${commandId}`, {
            method: "GET"
        }).then(async response => {
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Server error: ${response.status} - ${errorText}`);
            }
            return response.json()
        }).then(data => {
            currentCommand = data;
            renderRunButtonContainer();
        }).catch(err => {
            commandId = -1;
            location.hash=`no-commands`;
            renderNoCommands();
        });
    })
}

function renderCommandsList(withAnimation = false) {
    if (withAnimation) {
        document.getElementById("command-list").classList.add("gap-0-for-animation");
    }
    let listElem = document.getElementById("command-list");
    listElem.innerHTML = ""
    for (let i of commandsList) {
        let elem = document.createElement("li");
        elem.id = `command-btn-${i.id}`;
        if (withAnimation) {
            elem.classList.add("width-0-for-animation");
        }
        elem.innerHTML = `<button class="round-button" onclick="selectCommand(${i.id})">
                    <img width="80%" src="../static/vectors/arrow.svg" alt="Select command"/>
                </button>
                <p class="command-text command-name" onWheel="smoothHorizontalScroll(event)">
                    ${i.name}
                </p>`;
        listElem.appendChild(elem);
    }
    let elem = document.createElement("li");
    elem.id = `new-command-btn`;
    if (withAnimation) {
        elem.classList.add("width-0-for-animation");
    }
    elem.innerHTML = `<button class="round-button">
                    <img width="100%" src="../static/vectors/plus.svg" alt="Add new command"/>
                </button>
                <p class="command-text command-name" onWheel="smoothHorizontalScroll(event)">
                    Add new
                </p>`;
    elem.addEventListener("click", addNewCommand);
    listElem.appendChild(elem);
    selectButtonIcons(commandId);
    if (withAnimation) {
        setTimeout(() => {
            document.getElementById("command-list").classList.remove("gap-0-for-animation");
            for (elem of document.querySelectorAll(".width-0-for-animation")) {
                elem.classList.remove("width-0-for-animation");
            }
        }, 30);
    }
}

async function loadCommands() {
    try {
        let response = await fetch(`${apiBase}commands/`, {
            method: "GET"
        })
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
        commandsList = await response.json();
    } catch (err) {
        console.error('Ошибка:', err);
        showErrorPopup(
            'Ошибка загрузки команд',
            'Не удалось загрузить список команд. Попробуйте обновить страницу.',
            err.message
        );
    }
    return Promise
}

async function exportFiles() {
    let exportBtn
    let originalText
    try {
        exportBtn = document.getElementById('export-files-button');
        originalText = exportBtn.innerHTML;
        exportBtn.innerHTML = 'Exporting...';
        exportBtn.disabled = true;
        
        const response = await fetch(`${apiBase}files/download`, {
            method: 'GET'
        });
        
        if (response.ok) {
            const blob = await response.blob();
            
            const now = new Date();
            const dateStr = now.toISOString().slice(0, 10); // YYYY-MM-DD
            const timeStr = now.toTimeString().slice(0, 8).replace(/:/g, '-'); // HH-MM-SS
            const fileName = `all-files-${dateStr}-${timeStr}.zip`;
            
            saveFile(fileName, blob);
        } else {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
    } catch (err) {
        console.error('Error exporting files:', err);
        showErrorPopup(
            'Ошибка экспорта файлов',
            'Не удалось экспортировать файлы.',
            err.message
        );
    } finally {
        const exportBtn = document.getElementById('export-files-button');
        exportBtn.innerHTML = originalText;
        exportBtn.disabled = false;
    }
}

async function exportCommandFiles(commandId, commandName) {
    let tempBtn
    let originalDeleteBtn
    try {
        tempBtn = document.createElement('button');
        tempBtn.className = 'normal-button small-button';
        tempBtn.innerHTML = 'Exporting...';
        tempBtn.disabled = true;
        
        
        const response = await fetch(`${apiBase}commands/${commandId}/files/download`);

        if (response.ok) {
            const blob = await response.blob();

            const now = new Date();
            const dateStr = now.toISOString().slice(0, 10); // YYYY-MM-DD
            const timeStr = now.toTimeString().slice(0, 8).replace(/:/g, '-'); // HH-MM-SS
            const fileName = `command-${commandId}-files-${dateStr}-${timeStr}.zip`;

            saveFile(fileName, blob);
        } else {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
    } catch (err) {
        console.error('Error exporting command files:', err);
        showErrorPopup(
            'Ошибка экспорта файлов команды',
            `Не удалось экспортировать файлы для команды "${commandName}".`,
            err.message
        );
    } finally {
        const tempBtn = document.querySelector('.file-item button[disabled]');
        if (tempBtn) {
            tempBtn.parentNode.replaceChild(originalDeleteBtn, tempBtn);
        }
    }
}

async function downloadSingleFile(commandId, fileId, fileName) {
    try {
        const response = await fetch(`${apiBase}commands/${commandId}/files/${fileId}/download`);
        if (response.ok) {
            const blob = await response.blob();
            saveFile(fileName, blob);
        } else {
            throw new Error(`Failed to download file: ${response.status}`);
        }
    } catch (err) {
        console.error('Error downloading file:', err);
        showErrorPopup(
            'Ошибка скачивания файла',
            `Не удалось скачать файл "${fileName}".`,
            err.message
        );
    }
}

async function handleFileDownload(event) {
    const btn = event.target.closest('.download-file-btn');
    const fileId = btn.dataset.fileId;
    
    const fileItem = btn.closest('.file-item');
    const nameInput = fileItem.querySelector('.file-name-input');
    const fileName = nameInput.value.trim();
    const originalHTML = btn.innerHTML;
    try {
        btn.innerHTML = '...';
        btn.disabled = true;
        
        await downloadSingleFile(commandId, fileId, fileName);
        
    } catch (err) {
        console.error('Error downloading file:', err);
        showErrorPopup(
            'Ошибка скачивания файла',
            `Не удалось скачать файл "${fileName}".`,
            err.message
        );
    } finally {
        btn.innerHTML = originalHTML;
        btn.disabled = false;
    }
}

async function loadCommandFiles() {
    const filesList = document.getElementById("files-list");
    if (!filesList) return;
    
    try {
        const response = await fetch(`${apiBase}commands/${commandId}/files/`);
        if (response.ok) {
            const files = await response.json();
            renderFilesList(files);
        } else {
            const errorText = await response.text();
            filesList.innerHTML = `<p style="color: #E85D75;">Error loading files: ${response.status} - ${errorText}</p>`;
        }
    } catch (err) {
        console.error('Error loading files:', err);
        if (filesList) {
            filesList.innerHTML = '<p style="color: #E85D75;">Error loading files: Network error</p>';
        }
    }
}

function renderFilesList(files) {
    const filesList = document.getElementById("files-list");
    if (!filesList) return;
    
    if (files.length === 0) {
        filesList.innerHTML = '<p style="opacity: 0.7;">No files attached</p>';
        return;
    }

    filesList.innerHTML = '';
    
    const headerDiv = document.createElement('div');
    headerDiv.style.cssText = 'position: sticky;top: 0;background: #313034;z-index: 200;display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px; padding: 8px; border-radius: 5px;';
    headerDiv.innerHTML = `
        <span style="font-weight: 500; color: #6B7FD7;">Files (${files.length})</span>
        <button class="normal-button small-button export-command-files-btn" style="background-color: #6B7FD7;">
            Export All
        </button>
    `;
    filesList.appendChild(headerDiv);
    
    files.forEach(file => {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.style.cssText = 'display: flex; align-items: center; gap: 10px; margin-bottom: 8px; padding: 8px; background: rgba(49, 48, 52, 0.3); border-radius: 5px;';
        
        fileItem.innerHTML = `
            <input type="text" 
                   class="command-text command-name file-name-input" 
                   value="${file.name}" 
                   data-file-id="${file.id}"
                   style="flex: 1; margin: 0;"
                   spellcheck="false">
            <button class="normal-button small-button download-file-btn" 
                    data-file-id="${file.id}" 
                    style="background-color: #6B7FD7;">
                <img src="../static/vectors/arrow.svg" alt="Download" style="height: 1em; transform: rotate(-90deg);">
            </button>
            <button class="normal-button small-button delete-file-btn" 
                    data-file-id="${file.id}" 
                    style="background-color: #E85D75;">
                <img src="../static/vectors/delete.svg" alt="Delete" style="height: 1em;">
            </button>
        `;
        
        filesList.appendChild(fileItem);
    });

    const nameInputs = filesList.querySelectorAll('.file-name-input');
    if (nameInputs.length > 0) {
        nameInputs.forEach(input => {
            input.addEventListener('blur', handleFileNameEdit);
        });
    }

    const deleteBtns = filesList.querySelectorAll('.delete-file-btn');
    if (deleteBtns.length > 0) {
        deleteBtns.forEach(btn => {
            btn.addEventListener('click', handleFileDelete);
        });
    }
    
    const downloadBtns = filesList.querySelectorAll('.download-file-btn');
    if (downloadBtns.length > 0) {
        downloadBtns.forEach(btn => {
            btn.addEventListener('click', handleFileDownload);
        });
    }
    
    const exportBtn = filesList.querySelector('.export-command-files-btn');
    if (exportBtn) {
        exportBtn.addEventListener('click', () => {
            exportCommandFiles(commandId, currentCommand.name);
        });
    }
}

async function handleFileNameEdit(event) {
    const input = event.target;
    const fileId = input.dataset.fileId;
    const newName = input.value.trim();

    try {
        const response = await fetch(`${apiBase}commands/${commandId}/files/${fileId}`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                name: newName
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
    } catch (err) {
        console.error('Error updating file name:', err);
        showErrorPopup(
            'Ошибка обновления имени файла',
            'Не удалось обновить имя файла.',
            err.message
        );
        input.value = input.defaultValue;
    }
}

async function handleFileDelete(event) {
    const btn = event.target.closest('.delete-file-btn');
    const fileId = btn.dataset.fileId;

    try {
        const response = await fetch(`${apiBase}commands/${commandId}/files/${fileId}`, {
            method: 'DELETE'
        });

        if (response.ok) {
            loadCommandFiles();
        } else {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
    } catch (err) {
        console.error('Error deleting file:', err);
        showErrorPopup(
            'Ошибка удаления файла',
            'Не удалось удалить файл.',
            err.message
        );
    }
}

async function handleFileSelection(event) {
    const files = event.target.files;
    if (files.length === 0) return;

    // Показываем информацию о выбранных файлах
    const info = document.getElementById('selected-files-info');
    info.textContent = `${files.length} file(s) selected`;

    // Создаем FormData для отправки файлов
    const formData = new FormData();
    for (let i = 0; i < files.length; i++) {
        formData.append('files', files[i]);
    }

    try {
        const response = await fetch(`${apiBase}commands/${commandId}/files/`, {
            method: 'POST',
            body: formData
        });

        if (response.ok) {
            event.target.value = '';
            info.textContent = '';
            
            loadCommandFiles();
        } else {
            const errorText = await response.text();
            throw new Error(`Server error: ${response.status} - ${errorText}`);
        }
    } catch (err) {
        console.error('Error uploading files:', err);
        showErrorPopup(
            'Ошибка загрузки файлов',
            'Не удалось загрузить файлы.',
            err.message
        );
        info.textContent = '';
    }
}

window.addEventListener('beforeunload', () => {
    try {
        if (terminalWebsocket && terminalWebsocket.readyState === WebSocket.OPEN) {
            terminalWebsocket.close(4001, "terminal closed from frontend");
        }
    } catch (_) {}
});

// Функция для очистки обработчиков событий файлов
function cleanupFileHandlers() {
    const filesList = document.getElementById("files-list");
    if (!filesList) return;
    
    // Удаляем все обработчики событий
    const nameInputs = filesList.querySelectorAll('.file-name-input');
    nameInputs.forEach(input => {
        input.removeEventListener('blur', handleFileNameEdit);
    });
    
    const deleteBtns = filesList.querySelectorAll('.delete-file-btn');
    deleteBtns.forEach(btn => {
        btn.removeEventListener('click', handleFileDelete);
    });
    
    const downloadBtns = filesList.querySelectorAll('.download-file-btn');
    downloadBtns.forEach(btn => {
        btn.removeEventListener('click', handleFileDownload);
    });
    
    const exportBtn = filesList.querySelector('.export-command-files-btn');
    if (exportBtn) {
        exportBtn.removeEventListener('click', () => {
            exportCommandFiles(commandId, currentCommand.name);
        });
    }
}

function escapeHTML(str) {
    if (typeof str !== 'string') return str;
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
        .replace(/`/g, '&#96;');
}

function showErrorPopup(title, message, details = null) {
    const existingErrorPopup = document.getElementById('error-popup');
    if (existingErrorPopup) {
        document.body.removeChild(existingErrorPopup);
    }
    
    const popup = document.createElement('div');
    popup.id = 'error-popup';
    const safeTitle = escapeHTML(title);
    const safeMessage = escapeHTML(message);
    const safeDetails = details !== null ? escapeHTML(details) : null;
    popup.innerHTML = `
        <div class="popup-backdrop"></div>
        <div class="popup-content error-popup">
            <h2 class="error-title">${safeTitle}</h2>
            <div class="error-message">
                <p>${safeMessage}</p>
                ${safeDetails ? `<details class="error-details">
                    <summary>Подробности</summary>
                    <pre>${safeDetails}</pre>
                </details>` : ''}
            </div>
            <div class="popup-buttons">
                <button id="error-close-btn" class="normal-button">Закрыть</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(popup);
    
    setTimeout(() => {
        document.querySelector("#error-popup .popup-backdrop").classList.remove("hidden");
        document.querySelector("#error-popup .popup-content").classList.remove("hidden");
    }, 20);
    
    document.getElementById('error-close-btn').addEventListener('click', () => {
        closeErrorPopup();
    });
    
    document.querySelector("#error-popup .popup-backdrop").addEventListener('click', (e) => {
        if (e.target.classList.contains('popup-backdrop')) {
            closeErrorPopup();
        }
    });
    
    const escapeHandler = (e) => {
        if (e.key === 'Escape') {
            closeErrorPopup();
            document.removeEventListener('keydown', escapeHandler);
        }
    };
    document.addEventListener('keydown', escapeHandler);
}

function closeErrorPopup() {
    const popup = document.getElementById('error-popup');
    if (!popup) return;
    
    document.querySelector("#error-popup .popup-backdrop").classList.add("hidden");
    document.querySelector("#error-popup .popup-content").classList.add("hidden");
    
    setTimeout(() => {
        if (popup.parentNode) {
            popup.parentNode.removeChild(popup);
        }
    }, 300);
}


function importFiles(event) {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.zip';
    input.style.display = 'none';

    input.addEventListener('change', async () => {
        const file = input.files[0];
        if (!file) return;

        const formData = new FormData();
        formData.append('files', file);
        try {
            const response = await fetch(`${apiBase}files/upload`, {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                event.target.value = '';
            } else {
                const errorText = await response.text();
                throw new Error(`Server error: ${response.status} - ${errorText}`);
            }
        } catch (err) {
            console.error('Error uploading files:', err);
            showErrorPopup(
                'Ошибка загрузки файлов',
                'Не удалось загрузить файлы.',
                err.message
            );
            info.textContent = '';
        }
    });
    document.body.appendChild(input);
    input.click();
    document.body.removeChild(input);
}