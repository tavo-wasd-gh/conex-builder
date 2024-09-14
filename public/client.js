document.addEventListener("DOMContentLoaded", function() {
    const savedData = localStorage.getItem('editor_data');
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

    function openDialog() {
        dialog.style.display = "block";
        overlay.style.display = "block";
        menu.style.display = "none";
    }

    function closeDialog() {
        dialog.style.display = "none";
        overlay.style.display = "none";
        menu.style.display = "block";
    }

    document.getElementById("openDialogButton").addEventListener("click", openDialog);
    document.getElementById("cancelDialogButton").addEventListener("click", closeDialog);
    document.getElementById('title').addEventListener('change', saveEditorData);
    document.getElementById('slogan').addEventListener('change', saveEditorData);
});

function saveEditorData() {
    const banner = document.getElementById('banner').src;
    const title = document.getElementById('title').value;
    const slogan = document.getElementById('slogan').value;

    editor.save().then((editor_data) => {
        const dataToSave = {
            banner: banner,
            title: title,
            slogan: slogan,
            editor_data: editor_data
        };
        localStorage.setItem('editor_data', JSON.stringify(dataToSave));
        console.log('Editor data saved to localStorage');
    }).catch((error) => {
        console.error('Saving failed:', error);
    });
}

const directoryInput = document.getElementById('title');
const statusPopup = document.getElementById('status-popup');
const statusMessage = document.getElementById('status-message');

let typingTimeout;
let hideTimeout;

directoryInput.addEventListener('input', () => {
    clearTimeout(typingTimeout);
    typingTimeout = setTimeout(() => {
        const directoryName = directoryInput.value.trim();
        if (directoryName.length > 0) {
            const sanitizedDirectoryName = sanitizeDirectoryName(directoryName);
            checkDirectory(sanitizedDirectoryName);
        } else {
            hidePopup();
        }
    }, 500); // Debounce
});

function sanitizeDirectoryName(name) {
    return name
        .toLowerCase()
        .replace(/\s+/g, '-')
        .replace(/[^a-z0-9\-]/g, '');
}

function checkDirectory(name) {
    if (name.length < 4) {
        return;
    }
    fetch(`/api/directory/${encodeURIComponent(name)}`)
        .then(response => response.json())
        .then(data => {
            if (data.exists) {
                showPopup(`El sitio web conex.one/${name} ya existe`, 'exists');
            } else {
                showPopup(`Se publicará en conex.one/${name}`, 'available');
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
    const file = event.target.files[0];
    const reader = new FileReader();

    reader.onload = function (e) {
        const base64Image = e.target.result;
        document.getElementById('banner').src = base64Image;
        saveEditorData();
    };

    if (file) {
        reader.readAsDataURL(file);
    }
});
