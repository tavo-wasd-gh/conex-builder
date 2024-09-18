document.addEventListener("DOMContentLoaded", function() {
    loadLanguage('es');
    const savedData = localStorage.getItem('conex_data');
    if (savedData) {
        const parsedData = JSON.parse(savedData);
        console.log('Loaded parsedData:', parsedData);
        document.getElementById('title').value = parsedData.title || '';
        document.getElementById('slogan').value = parsedData.slogan || '';
        document.getElementById('banner').src = parsedData.banner || '/static/svg/banner.svg';
    }

    const dialog = document.getElementById("dialog");
    const overlay = document.getElementById("overlay");
    const menu = document.getElementById("floatingButtons");
    const checkoutErrorMessage = document.getElementById("checkout-error-message");

    function openDialog() {
        checkoutErrorMessage.style.display = "none";
        dialog.style.display = "block";
        overlay.style.display = "block";
        menu.style.display = "none";
    }

    function closeDialog() {
        checkoutErrorMessage.style.display = "none";
        dialog.style.display = "none";
        overlay.style.display = "none";
        menu.style.display = "flex";
    }

    document.getElementById("openDialogButton").addEventListener("click", openDialog);
    document.getElementById("cancelDialogButton").addEventListener("click", closeDialog);
    document.getElementById('title').addEventListener('change', saveEditorData);
    document.getElementById('slogan').addEventListener('change', saveEditorData);
    document.getElementById('buyModeButton').addEventListener('click', buyMode);
    document.getElementById('editModeButton').addEventListener('click', editModeWrapper);
    document.getElementById("continueToEditMode").addEventListener('click', function() {
        editMode(document.getElementById("dashEditInput").value);
    });
    document.getElementById('dashButton').addEventListener('click', dashboardMode);
});

function saveEditorData() {
    const banner = document.getElementById('banner').src;
    const title = document.getElementById('title').value;
    const slogan = document.getElementById('slogan').value;

    editor.save().then((editor_data) => {
        const dataToSave = {
            directory: sanitizeDirectoryTitle(title),
            banner: banner || '/static/svg/banner.svg',
            title: title,
            slogan: slogan,
            editor_data: editor_data
        };
        localStorage.setItem('conex_data', JSON.stringify(dataToSave));
        console.log('Editor data saved to localStorage');
    }).catch((error) => {
        console.error('Saving failed:', error);
    });
}

let typingTimeout;
let hideTimeout;
const directoryInput = document.getElementById('title');
directoryInput.addEventListener('input', () => {
    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => {
        const directoryTitle = directoryInput.value.trim();
        if (directoryTitle.length > 0) {
            const directory = sanitizeDirectoryTitle(directoryTitle);
            checkDirectory(directory);
        } else {
            hidePopup();
        }
    }, 500); // Debounce
});

const directoryBuyInput = document.getElementById('dashEditInput');
directoryBuyInput.addEventListener('input', () => {
    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => {
        const directoryTitle = directoryBuyInput.value.trim();
        if (directoryTitle.length > 0) {
            const directory = sanitizeDirectoryTitle(directoryTitle);
            checkDirectoryReverse(directory);
        } else {
            hidePopup();
        }
    }, 500); // Debounce
});

function sanitizeDirectoryTitle(title) {
    return title
        .toLowerCase()
        .replace(/\s+/g, '-')
        .normalize('NFD')
        .replace(/[\u0300-\u036f]/g, '')
        .replace(/[^a-z0-9\-]/g, '');
}

function checkDirectory(directory) {
    if (directory.length < 4) {
        return;
    }
    if (directory.length > 35) {
        showPopup(`El título no puede exceder los 35 caracteres`, 'exists');
        return;
    }
    fetch(`/api/directory/${encodeURIComponent(directory)}`)
        .then(response => response.json())
        .then(data => {
            if (data.exists) {
                showPopup(`El sitio web conex.one/${directory} ya existe`, 'exists');
            } else {
                showPopup(`Se publicará en conex.one/${directory}`, 'available');
            }
        })
        .catch(error => {
            console.error('Error checking directory:', error);
            showPopup('Error checking directory.', 'exists');
        });
}

function checkDirectoryReverse(directory) {
    if (directory.length < 4) {
        return;
    }
    if (directory.length > 35) {
        showPopup(`El título no puede exceder los 35 caracteres`, 'exists');
        return;
    }
    fetch(`/api/directory/${encodeURIComponent(directory)}`)
        .then(response => response.json())
        .then(data => {
            if (data.exists) {
                showPopup(`Editar el sitio web conex.one/${directory}`, 'available');
            } else {
                showPopup(`El sitio conex.one/${directory} no está registrado`, 'exists');
            }
        })
        .catch(error => {
            console.error('Error checking directory:', error);
            showPopup('Error checking directory.', 'exists');
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
    popup.classList.remove('show');
    setTimeout(() => {
        popup.classList.remove(status);
    }, 100);
}

document.getElementById('imageUpload').addEventListener('change', function (event) {
    const savedData = localStorage.getItem('conex_data');
    const parsedData = savedData ? JSON.parse(savedData) : null;
    const directory = parsedData?.directory || "temp";
    const file = event.target.files[0];

    if (file) {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('directory', directory);

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
        });
    }
});

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
    document.getElementById("dashOverlay").style.display = "none";
    document.getElementById("dashDialog").style.display = "none";

    const dashboard = document.querySelector('.dashboard');

    dashboard.style.display = 'flex';
    setTimeout(() => {
        dashboard.style.transition = 'opacity 0.5s ease';
        dashboard.style.opacity = '1';
    }, 10);
}

function buyMode() {
    document.getElementById("openDialogButton").style.display = "block";
    document.getElementById("updateSiteButton").style.display = "none";
    const dashboard = document.querySelector('.dashboard');
    dashboard.style.transition = 'opacity 0.5s ease';
    dashboard.style.opacity = '0';

    setTimeout(() => {
        dashboard.style.display = 'none';
    }, 500);
}

async function editMode(dir) {
    const success = await loadEditorData(dir);
    if (!success) {
        console.error("Data could not be loaded, aborting UI changes");
        return;
    }
    document.getElementById("openDialogButton").style.display = "none";
    document.getElementById("updateSiteButton").style.display = "block";
    const dashboard = document.querySelector('.dashboard');
    dashboard.style.transition = 'opacity 0.5s ease';
    dashboard.style.opacity = '0';

    setTimeout(() => {
        dashboard.style.display = 'none';
    }, 500);
}

function editModeWrapper() {
    document.getElementById("dashOverlay").style.display = "block";
    document.getElementById("dashDialog").style.display = "block";
    document.getElementById("dashEditInput").style.display = "block";
}

function loadEditorData(dirname) {
    return new Promise((resolve, reject) => {
        fetch(`/api/fetch/${dirname}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                // Populate the data into the form elements
                document.getElementById('banner').src = data.banner || '/static/svg/banner.svg';
                document.getElementById('title').value = data.title || '';
                document.getElementById('slogan').value = data.slogan || '';

                // Load editor data if available
                if (data.editor_data && typeof editor !== 'undefined') {
                    editor.render({
                        blocks: data.editor_data.blocks || []
                    });
                }

                console.log('Editor data loaded successfully');
                resolve(true); // Resolve with success
            })
            .catch(error => {
                console.error('Fetching and loading data failed:', error);
                resolve(false); // Resolve with failure, but not reject to avoid unhandled errors
            });
    });
}
