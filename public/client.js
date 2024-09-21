const PayPalSDK = "https://sandbox.paypal.com/sdk/js?client-id=AcCW43LI1S6lLQgtLkF4V8UOPfmXcqXQ8xfEl41hRuMxSskR2jkWNwQN6Ab1WK7E2E52GNaoYBHqgIKd&components=buttons&enable-funding=card&disable-funding=paylater,venmo"

const EditorJSComponents = [
    "https://cdn.jsdelivr.net/npm/@editorjs/header@latest",
    "https://cdn.jsdelivr.net/npm/@editorjs/image@latest",
    "https://cdn.jsdelivr.net/npm/@editorjs/list@latest",
    "https://cdn.jsdelivr.net/npm/@editorjs/quote@latest",
    "https://cdn.jsdelivr.net/npm/@editorjs/code@latest",
    "https://cdn.jsdelivr.net/npm/@editorjs/table@latest",
    "https://cdn.jsdelivr.net/npm/@editorjs/editorjs@latest",
];

let typingTimeout;
let hideTimeout;
let editor;

document.addEventListener("DOMContentLoaded", function () {
    const dialog = document.getElementById("dialog");
    const overlay = document.getElementById("overlay");
    const floatingButtons = document.getElementById("floatingButtons");
    const checkoutDialog = document.getElementById("checkoutDialog");
    const editDialog = document.getElementById("editDialog");
    const buyDialog = document.getElementById("buyDialog");
    const updateContentDialog = document.getElementById("updateContentDialog");

    checkoutDialog.style.display = "none";
    editDialog.style.display = "none";
    buyDialog.style.display = "none";
    updateContentDialog.style.display = "none";
    loadLanguage('es');

    Promise.all(EditorJSComponents.map(src => loadScript(src))).then(() => {
        return loadScript("/editor.js");
    }).then(() => {
        loadEditorState();
    }).catch(err => console.error("Error loading editor:", err));

    loadScript(PayPalSDK).then(() => {
        return loadScript("/paypal.js");
    }).catch(err => console.error(err));

    initializeEventListeners();
});

function loadScript(src, async = true) {
    return new Promise((resolve, reject) => {
        const script = document.createElement('script');
        script.src = src;
        script.async = async;
        script.onload = () => resolve();
        script.onerror = () => reject(new Error(`Failed to load script: ${src}`));
        document.head.appendChild(script);
    });
}

function loadEditorState() {
    const savedData = localStorage.getItem('conex_data');
    const holderElement = document.getElementById('editorjs');
    holderElement.innerHTML = '';

    if (savedData) {
        const parsedData = JSON.parse(savedData);
        console.log('Loaded parsedData:', parsedData);
        document.getElementById('title').innerText = parsedData.title || '';
        document.getElementById('slogan').value = parsedData.slogan || '';
        document.getElementById('banner').src = parsedData.banner || '/static/svg/banner.svg';
        const disclaimers = document.querySelectorAll(".localstorage-exists");
        disclaimers.forEach((element) => {
            element.style.display = "block";
        });

        initializeEditor(parsedData);

        document.getElementById('continueEditingModeButton').style.display = "block";
        fetch(`./lang/es.json`)
            .then(response => response.json())
            .then(translations => {
                const translatedText = translations['continueEditingModeButton'];
                const continueEditingModeButton = document.getElementById('continueEditingModeButton');
                if (parsedData.title && translatedText) {
                    continueEditingModeButton.innerText = `${translatedText}: ${parsedData.title}`;
                } else {
                    continueEditingModeButton.innerText = translatedText;
                }
            })
            .catch(error => console.error('Error loading translation file:', error));
    } else {
        document.getElementById('title').innerText = '';
        document.getElementById('slogan').value = '';
        document.getElementById('banner').src = '/static/svg/banner.svg';
        document.getElementById("continueEditingModeButton").style.display = "none";
        const disclaimers = document.querySelectorAll(".localstorage-exists");
        disclaimers.forEach((element) => {
            element.style.display = "none";
        });

        initializeEditor();
    }
}

function initializeEventListeners() {
    // Prevent ENTER key in slogan element
    document.getElementById('slogan').addEventListener('keydown', function(event) {
        if (event.key === 'Enter') {
            event.preventDefault();
        }
    });

    document.getElementById("buyButton").addEventListener("click", () => openDialog(checkoutDialog));
    document.getElementById("editButton").addEventListener("click", () => openDialog(updateContentDialog));
    document.getElementById("closeDialogButton").addEventListener("click", () => closeDialog());

    // // Editor save
    // document.getElementById('title').addEventListener('change', saveEditorData);
    document.getElementById('slogan').addEventListener('change', saveEditorData);

    // // Mode switching
    document.getElementById('buyModeButton').addEventListener('click', openBuyModeDialog);
    document.getElementById('editModeButton').addEventListener('click', openEditModeDialog);
    document.getElementById("continueToBuyModeButton").addEventListener('click', async () => {
        const directory = sanitizeDirectoryTitle(document.getElementById("buyModeDirectoryInput").value);
        const exists = await checkDirectoryExists(directory);
        if (exists) {
            document.getElementById("checkdir-error-message").style.display = "block";
            document.getElementById("checkdir-error-message").innerHTML = `El sitio https://conex.one/${directory} ya existe.`;
        } else {
            document.getElementById("checkdir-error-message").style.display = "none";
            buyMode(directory);
        }
    });

    document.getElementById("continueToEditModeButton").addEventListener('click', () =>
        editMode(extractSitePath(document.getElementById("editModeDirectoryInput").value))
    );
    document.getElementById('continueEditingModeButton').addEventListener('click', continueMode);

    document.getElementById('dashButton').addEventListener('click', dashboardMode);
    document.getElementById('uploadBannerBtn').addEventListener('change', handleImageUpload);

    document.getElementById('requestChangesButton').addEventListener('click', async () => {
        const button = document.getElementById('requestChangesButton');
        button.disabled = true;
        button.classList.add('disabled');

        await updateSiteRequest();

        setTimeout(() => {
            button.disabled = false;
            button.classList.remove('disabled');
        }, 3000);
    });

    document.getElementById('confirmChangesButton').addEventListener('click', function() {
        successElement = document.getElementById('update-success-message');
        errorElement = document.getElementById('update-error-message');

        const codeInput = document.getElementById('updateContentCodeInput').value;
        if (codeInput.length === 6 && !isNaN(codeInput)) {
            document.getElementById('updateContentCodeInput').value = '';
            errorElement.style.display = "none"
            updateSiteConfirm(codeInput);
        } else {
            successElement.style.display = "none"
            errorElement.style.display = "block"
            errorElement.innerHTML = "El código es un pin numérico de 6 dígitos.";
            console.error('Invalid code. Please enter a 6-digit number.');
        }
    });

    const titleElement = document.getElementById('buyModeDirectoryInput');
    titleElement.addEventListener('input', debounce(function() {
        const directory = titleElement.value.trim();
        if (directory.length > 0) {
            validateDirectory(directory);
        }
    }, 500));  // 500ms debounce
}

function openDialog(content) {
    dialog.style.display = "block";
    content.style.display = "block";
    overlay.style.display = "block";
    floatingButtons.style.display = "none";
}

function debounce(func, delay) {
    let timeout;
    return function(...args) {
        clearTimeout(timeout);
        timeout = setTimeout(() => func.apply(this, args), delay);
    };
}

function closeDialog() {
    checkoutDialog.style.display = "none";
    editDialog.style.display = "none";
    buyDialog.style.display = "none";
    updateContentDialog.style.display = "none";
    dialog.style.display = "none";
    overlay.style.display = "none";
    floatingButtons.style.display = "flex";
    document.getElementById('checkout-error-message').style.display = "none";
    document.getElementById('update-error-message').style.display = "none";
}

function saveEditorData() {
    const titleValue = document.getElementById('title').innerText.trim();
    const dataToSave = {
        banner: document.getElementById('banner').src || '/static/svg/banner.svg',
        title: document.getElementById('title').innerText,
        slogan: document.getElementById('slogan').value,
        directory: sanitizeDirectoryTitle(titleValue)
    };
    editor.save().then((editor_data) => {
        dataToSave.editor_data = editor_data;
        localStorage.setItem('conex_data', JSON.stringify(dataToSave));
        console.log('Editor data saved to localStorage');
    }).catch((error) => {
        console.error('Saving failed:', error);
    });
}

function validateDirectory(directory) {
    successMessageElement = document.getElementById('checkdir-success-message');
    errorMessageElement = document.getElementById('checkdir-error-message');
    successMessageElement.textContent = '';
    errorMessageElement.textContent = '';

    if (!validateDirectoryLength(directory)) {
        successMessageElement.style.display = "none";
        errorMessageElement.style.display = "block";
        errorMessageElement.textContent = 'Directory name must be between 4 and 35 characters.';
        return;
    }

    directory = sanitizeDirectoryTitle(directory)
    checkDirectoryExists(directory)
        .then(exists => {
            if (exists) {
                successMessageElement.style.display = "none";
                errorMessageElement.style.display = "block";
                errorMessageElement.textContent = `El sitio https://conex.one/${directory} ya existe.`;
            } else {
                successMessageElement.style.display = "block";
                errorMessageElement.style.display = "none";
                successMessageElement.textContent = `Se publicará en https://conex.one/${directory}`;
            }
        })
        .catch(() => {
            successMessageElement.style.display = "none";
            errorMessageElement.style.display = "block";
            errorMessageElement.textContent = 'Error occurred while checking the directory.';
        });
}

function sanitizeDirectoryTitle(title) {
    return title
        .toLowerCase()
        .replace(/\s+/g, '-')
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .replace(/[^a-z0-9\-]/g, '');
}

function validateDirectoryLength(directory) {
    if (directory.length < 4 || directory.length > 35) {
        showPopup('El título debe tener entre 4 y 35 caracteres', 'exists');
        return false;
    }
    return true;
}

function checkDirectoryExists(directory) {
    return fetch(`/api/directory/${encodeURIComponent(directory)}`)
        .then(response => response.json())
        .then(data => data.exists)
        .catch(error => {
            console.error('Error checking directory:', error);
            return false;
        });
}

function showPopup(message, status) {
    const popup = document.querySelector('.status-popup');

    if (hideTimeout) {
        clearTimeout(hideTimeout);
    }

    const existingCloseButton = popup.querySelector('.close-popup');
    if (existingCloseButton) {
        existingCloseButton.remove();
    }

    const closeButton = document.createElement('span');
    closeButton.classList.add('close-popup');
    closeButton.innerHTML = '&times;';
    closeButton.addEventListener('click', () => {
        hidePopup(popup, status);
    });

    popup.innerHTML = message;
    popup.appendChild(closeButton);
    popup.classList.remove('exists', 'available');
    popup.classList.add(status);
    popup.classList.add('show');

    hideTimeout = setTimeout(() => {
        hidePopup(popup, status);
    }, 5000);
}

function hidePopup(popup, status) {
    if (!popup) {
        return;
    }
    popup.classList.remove('show');
    setTimeout(() => {
        popup.classList.remove(status);
    }, 100);
}

function handleImageUpload() {
    const uploadButton = document.getElementById('uploadBannerBtn');
    const imageIcon = document.querySelector('.tool-button img');
    const loader = document.querySelector('.loader');

    const savedData = localStorage.getItem('conex_data');
    const parsedData = savedData ? JSON.parse(savedData) : null;
    const directory = parsedData?.directory || "temp";
    const file = event.target.files[0];

    if (file) {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('directory', directory);

        uploadButton.disabled = true;
        loader.style.display = 'inline-block';
        imageIcon.style.display = 'none';

        fetch('/api/upload', {
            method: 'POST',
            body: formData,
        }).then(response => response.json()).then(data => {
            if (data && data.file && data.file.url) {
                document.getElementById('banner').src = data.file.url;
                saveEditorData();
            } else {
                console.error('Error: Invalid response format', data);
            }
        }).catch(error => {
            console.error('Error uploading the image:', error);
        }).finally(() => {
            uploadButton.disabled = false;
            loader.style.display = 'none';
            imageIcon.style.display = 'inline-block';
        });
    }
}

function loadLanguage(lang) {
    fetch(`./lang/${lang}.json`)
        .then(response => response.json())
        .then(translations => {
            // Find all elements with a 'data-translate' attribute
            document.querySelectorAll('[data-translate]').forEach(element => {
                const translationKey = element.getAttribute('data-translate');

                // Check if the element is an input field (update placeholder)
                if (element.tagName.toLowerCase() === 'input' || element.tagName.toLowerCase() === 'textarea') {
                    element.placeholder = translations[translationKey];
                } else {
                    // Update text content for non-input elements
                    element.innerText = translations[translationKey];
                }
            });
        })
        .catch(error => console.error('Error loading language file:', error));
}

function dashboardMode() {
    loadEditorState();
    const dashboard = document.getElementById('dashboard');
    dashboard.style.display = 'flex';
    dashboard.style.opacity = '0';
    setTimeout(() => {
        dashboard.style.transition = 'opacity 0.5s ease';
        dashboard.style.opacity = '1';
    }, 10);
}

function buyMode() {
    localStorage.removeItem('conex_data');
    const dataToSave = {
        title: document.getElementById('buyModeDirectoryInput').value.trim(),
    };
    localStorage.setItem('conex_data', JSON.stringify(dataToSave));
    loadEditorState();

    closeDialog();

    document.getElementById('checkout-success-message').style.display = "none";
    document.querySelector("#paypal-button-container").style.display = "block";
    document.getElementById("buyButton").style.display = "block";
    document.getElementById("editButton").style.display = "none";

    const dashboard = document.getElementById('dashboard');
    dashboard.style.transition = 'opacity 0.5s ease';
    dashboard.style.opacity = '0';

    setTimeout(() => {
        dashboard.style.display = 'none';
    }, 500);
}

function continueMode() {
    const savedData = localStorage.getItem('conex_data');
    if (savedData) {
        const parsedData = JSON.parse(savedData);
        if (parsedData.directory) {
            checkDirectoryExists(parsedData.directory).then(exists => {
                if (exists) {
                    editMode(parsedData.directory);
                } else {
                    continueBuyMode();
                }
            });
        } else {
            continueBuyMode();
        }
    } else {
        buyMode();
    }
}

function continueBuyMode() {
    document.getElementById("buyButton").style.display = "block";
    document.getElementById("editButton").style.display = "none";
    const dashboard = document.getElementById('dashboard');
    dashboard.style.transition = 'opacity 0.5s ease';
    dashboard.style.opacity = '0';

    setTimeout(() => {
        dashboard.style.display = 'none';
    }, 500);
}

function openBuyModeDialog() {
    document.getElementById("checkdir-error-message").style.display = "none";
    document.getElementById("checkdir-success-message").style.display = "none";
    overlay.style.display = "block";
    dialog.style.display = "block";
    buyDialog.style.display = "block";
}

function openEditModeDialog() {
    overlay.style.display = "block";
    dialog.style.display = "block";
    editDialog.style.display = "block";
}

async function editMode(dir) {
    const errorMessageElement = document.getElementById('edit-error-message');
    const conexData = JSON.parse(localStorage.getItem('conex_data'));

    if (conexData?.directory === dir) {
        console.log("Directory already loaded, skipping fetch.");
    } else {
        const success = await fetchAndStoreData(dir);
        if (!success) {
            errorMessageElement.innerHTML = "No se pudo cargar el sitio, asegúrate que estás digitando el enlace correcto.";
            errorMessageElement.style.display = "block";
            console.error("Data could not be loaded, aborting UI changes");
            return;
        }
    }

    closeDialog();
    errorMessageElement.style.display = "none";
    document.getElementById("buyButton").style.display = "none";
    document.getElementById("editButton").style.display = "block";
    const dashboard = document.getElementById('dashboard');
    dashboard.style.transition = 'opacity 0.5s ease';
    dashboard.style.opacity = '0';

    setTimeout(() => {
        dashboard.style.display = 'none';
    }, 500);
}

async function fetchAndStoreData(directoryName) {
    try {
        const response = await fetch(`/api/fetch/${encodeURIComponent(directoryName)}`);
        if (!response.ok) {
            throw new Error(`Failed to fetch data for directory: ${directoryName}`);
        }

        const data = await response.json();
        localStorage.setItem('conex_data', JSON.stringify(data));
        console.log('Data fetched and stored in localStorage:', data);

        loadEditorState();
        return true;
    } catch (error) {
        console.error('Error fetching and storing data:', error);
        return false;
    }
}

function updateSiteRequest() {
    const conexData = JSON.parse(localStorage.getItem('conex_data'));
    const directory = conexData?.directory;
    successElement = document.getElementById('update-success-message');
    errorElement = document.getElementById('update-error-message');

    fetch('/api/update', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ directory: directory })
    })
    .then(response => {
        if (response.status === 200) {
            successElement.style.display = "block"
            errorElement.style.display = "none"
            successElement.innerHTML = "Se envió el código de autenticación de 6 dígitos a su correo electrónico.";
        } else {
            successElement.style.display = "none"
            errorElement.style.display = "block"
            errorElement.innerHTML = "Error enviando el código de confirmación a su correo, recuerde que puede solicitar el código solamente una vez cada minuto.";
        }
    })
}

function updateSiteConfirm(code) {
    const conexData = JSON.parse(localStorage.getItem('conex_data'));
    const directory = conexData?.directory;
    const editorData = conexData?.editor_data;
    const slogan = conexData?.slogan;

    successElement = document.getElementById('update-success-message');
    errorElement = document.getElementById('update-error-message');

    if (!directory || !editorData) {
        console.error('Directory or editor_data not found in localStorage');
        return;
    }

    fetch('/api/confirm', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            directory: directory,
            auth_code: code,
            slogan: slogan,
            editor_data: editorData
        })
    })
    .then(response => {
        if (response.status === 200) {
            successElement.style.display = "block"
            errorElement.style.display = "none"
            successElement.innerHTML = "Se actualizó correctamente la información de su sitio, los cambios deberían verse reflejados en menos de 24 horas.";
        } else {
            successElement.style.display = "none"
            errorElement.style.display = "block"
            errorElement.innerHTML = "Error actualizando su sitio, por favor vuelva a intentarlo más tarde.";
        }
    })
}

function extractSitePath(url) {
    if (!url.includes("conex.one")) {
        return url;
    }
    const cleanUrl = url.replace(/^(https?:\/\/)?(www\.)?/, '').replace(/#.*$/, '');
    const match = cleanUrl.match(/^conex\.one\/([^\/?#]+)\/?/);
    return match ? match[1] : null;
}
