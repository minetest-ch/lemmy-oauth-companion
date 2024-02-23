let created = false;

function createButtons(parentEl) {
    created = true;
    parentEl.className = "col-12";
    parentEl.append(document.createElement("hr"));

    let b = document.createElement("a");
    b.className = "btn btn-primary w-100";
    b.innerHTML = `<img src="/oauth-login/assets/github.png" height="24" width="24"/> Login with Github`;
    b.href = "/oauth-login/github";
    b.style["margin-bottom"] = "5px";
    parentEl.append(b);

    b = document.createElement("a");
    b.className = "btn btn-primary w-100";
    b.innerHTML = `<img src="/oauth-login/assets/contentdb.png" height="24" width="24"/> Login with ContentDB`;
    b.href = "/oauth-login/cdb";
    b.style["margin-bottom"] = "5px";
    parentEl.append(b);

    b = document.createElement("a");
    b.className = "btn btn-primary w-100";
    b.innerHTML = `<img src="/oauth-login/assets/default_mese_crystal.png" height="24" width="24"/> Login with Mesehub`;
    b.href = "/oauth-login/mesehub";
    b.style["margin-bottom"] = "5px";
    parentEl.append(b);
}

function findButtonContainer() {
    const elements = document.getElementsByClassName("btn-secondary");
    for (let i=0; i<elements.length; i++) {
        let el = elements[i];
        if (el.type == "submit") {
            el.className += " w-100";
            return el.parentElement;
        }
    }
}

// dirty hack to inject login-buttons until we have proper oauth support (let me know if you have a better idea ¯\_(ツ)_/¯)
function checkLoginpage() {
    if (location.pathname != "/login") {
        created = false;
        // not on the login page
        return;
    }

    if (created) {
        // already created
        return;
    }

    const c = findButtonContainer();
    if (c) {
        createButtons(c);
    }
}

document.addEventListener("DOMContentLoaded", checkLoginpage);
window.addEventListener("click", checkLoginpage);
setInterval(checkLoginpage, 1000);