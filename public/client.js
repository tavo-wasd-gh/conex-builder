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

    checkoutDialog.style.display = "none";
    editDialog.style.display = "none";
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
        document.getElementById('title').value = parsedData.title || '';
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
        document.getElementById('title').value = '';
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
    document.getElementById("closeDialogButton").addEventListener("click", () => closeDialog());

    // // Editor save
    document.getElementById('title').addEventListener('change', saveEditorData);
    document.getElementById('slogan').addEventListener('change', saveEditorData);

    // // Mode switching
    // document.getElementById('buyModeButton').addEventListener('click', openBuyModeDialog);
    document.getElementById('buyModeButton').addEventListener('click', buyMode);
    document.getElementById('editModeButton').addEventListener('click', openEditModeDialog);
    document.getElementById("continueToEditModeButton").addEventListener('click', () =>
        editMode(document.getElementById("editModeDirectoryInput").value)
    );
    document.getElementById('continueEditingModeButton').addEventListener('click', continueMode);

    document.getElementById('dashButton').addEventListener('click', dashboardMode);
    document.getElementById('uploadBannerBtn').addEventListener('change', handleImageUpload);

    const titleElement = document.getElementById('title');
    if (titleElement) {
        setupDirectoryInput(titleElement);
    }
}

function openDialog(content) {
    dialog.style.display = "block";
    content.style.display = "block";
    overlay.style.display = "block";
    floatingButtons.style.display = "none";
}

function closeDialog() {
    checkoutDialog.style.display = "none";
    editDialog.style.display = "none";
    // buyDialog.style.display = "none";
    dialog.style.display = "none";
    overlay.style.display = "none";
    floatingButtons.style.display = "flex";
    document.getElementById('checkout-error-message').style.display = "none";
}

function saveEditorData() {
    const titleValue = document.getElementById('title').value.trim();
    const dataToSave = {
        banner: document.getElementById('banner').src || '/static/svg/banner.svg',
        title: document.getElementById('title').value,
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

function setupDirectoryInput(inputElement, debounceTime = 500) {
    inputElement.addEventListener('input', () => {
        clearTimeout(typingTimeout);

        typingTimeout = setTimeout(() => {
            const inputValue = inputElement.value.trim();

            if (inputValue.length > 0) {
                const sanitizedValue = sanitizeDirectoryTitle(inputValue);  // Sanitize the input value
                checkDirectory(sanitizedValue);
            } else {
                hidePopup();
            }
        }, debounceTime);
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

function checkDirectory(directory) {
    if (!validateDirectoryLength(directory)) return;
    fetchDirectoryStatus(directory, 'exists', 'available', 'El sitio web ya existe', 'Se publicará en');
}

function validateDirectoryLength(directory) {
    if (directory.length < 4 || directory.length > 35) {
        showPopup('El título debe tener entre 4 y 35 caracteres', 'exists');
        return false;
    }
    return true;
}

function fetchDirectoryStatus(directory, failureStatus, successStatus, failureMessage, successMessage) {
    checkDirectoryExists(directory).then(exists => {
        const message = exists ? `${failureMessage} conex.one/${directory}` : `${successMessage} conex.one/${directory}`;
        const status = exists ? failureStatus : successStatus;
        showPopup(message, status);
    });
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
    loadEditorState();

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
    closeDialog();
    const success = await fetchAndStoreData(dir);
    if (!success) {
        console.error("Data could not be loaded, aborting UI changes");
        return;
    }

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
