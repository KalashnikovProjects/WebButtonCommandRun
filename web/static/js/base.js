let sumDelta = 0
let startLeft = 0
let targetLeft = 0

function smoothHorizontalScroll(e) {
    let el = e.target
    if (el.scrollLeft === targetLeft) sumDelta = 0
    if (sumDelta === 0) startLeft = el.scrollLeft
    sumDelta -= e.deltaY
    const maxScrollLeft = el.scrollWidth - el.clientWidth;
    const left = Math.min(maxScrollLeft, Math.max(0, startLeft - sumDelta));
    targetLeft = left
    el.scrollTo({left, behavior: "smooth"})
    e.preventDefault();
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
    document.body.classList.add("terminal-opened")
    // TODO
}

function closeTerminal(event) {
    document.body.classList.remove("terminal-opened")
    // TODO
}

function initPage() {
    let openMenuButton = document.getElementById("open-menu-button")
    openMenuButton.addEventListener("click", toggleMenu)

    let runCommandButton = document.getElementById("big-run-button")
    runCommandButton.addEventListener("click", runCommand)

    let closeTerminalButton = document.getElementById("close-button")
    closeTerminalButton.addEventListener("click", closeTerminal)
}

initPage()
