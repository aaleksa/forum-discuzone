document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('post-form');
    const imageInput = document.getElementById('post-image');
    const previewContainer = document.getElementById('image-preview');
    const primarySelector = document.getElementById('primary-image-selector');
    const primaryIndexInput = document.getElementById('primary-image-index');

    if (!form || !imageInput || !previewContainer || !primarySelector || !primaryIndexInput) return;

    let currentFiles = [];
    let currentPreviews = [];

    // Hide the primary image selector while it is empty
    primarySelector.style.display = 'none';

    imageInput.addEventListener('change', function (e) {
        const newFiles = Array.from(e.target.files);
        if (newFiles.length === 0) return;

        primarySelector.style.display = 'flex';

        const startIndex = currentFiles.length;
        currentFiles.push(...newFiles);

        newFiles.forEach((file, relativeIndex) => {
            const absoluteIndex = startIndex + relativeIndex;
            if (!file.type.match(/^image\/(jpeg|jpg|png|gif)$/)) return;

            const reader = new FileReader();
            reader.onload = function (e) {
                const previewItem = document.createElement('div');
                previewItem.className = 'preview-item';

                const radio = document.createElement('input');
                radio.type = 'radio';
                radio.name = 'primaryImg';
                radio.id = `primary-img-${absoluteIndex}`;
                radio.value = absoluteIndex;

                // Вибираємо перше додане зображення як первинне, якщо ще немає вибору
                if (primaryIndexInput.value === '0' && currentPreviews.length === 0) {
                    radio.checked = true;
                    primaryIndexInput.value = absoluteIndex.toString();
                }

                radio.addEventListener('change', function () {
                    primaryIndexInput.value = this.value;
                    document.querySelectorAll('.preview-item img').forEach(img => {
                        img.style.borderColor = 'transparent';
                    });
                    this.nextElementSibling.style.borderColor = '#4285f4';
                });

                const img = document.createElement('img');
                img.src = e.target.result;
                img.alt = file.name;
                img.classList.add('hover-zoom');
                
                previewItem.append(radio, img);
                previewContainer.appendChild(previewItem);
                currentPreviews.push(previewItem);

                const selectorRadio = radio.cloneNode(true);
                selectorRadio.id = `primary-img-${absoluteIndex}-selector`;

                const selectorLabel = document.createElement('label');
                selectorLabel.htmlFor = selectorRadio.id;
                selectorLabel.textContent = `Image ${absoluteIndex + 1}`;

                primarySelector.appendChild(selectorRadio);
                primarySelector.appendChild(selectorLabel);

                if (radio.checked) {
                    img.style.borderColor = '#4285f4';
                }
            };
            reader.readAsDataURL(file);
        });

        imageInput.value = '';
    });

    form.addEventListener('submit', function (e) {
        // If you need a check, you can add it here. If the images are optional, then we let them go.        
        e.preventDefault();
        const formData = new FormData(form);

        currentFiles.forEach(file => {
            formData.append('images[]', file);
        });

        fetch(form.action, {
            method: form.method,
            body: formData,
        })
            .then(res => {
                if (res.ok) {
                    window.location.href = '/posts';
                } else {
                    alert('Failed to upload');
                }
            })
            .catch(err => {
                console.error(err);
                alert('Error occurred');
            });
    });

    const clearButton = document.createElement('button');
    clearButton.textContent = 'Clear all images';
    clearButton.type = 'button';
    clearButton.className = 'btn-clear';
    clearButton.addEventListener('click', () => {
        previewContainer.innerHTML = '';
        primarySelector.innerHTML = '';
        primarySelector.style.display = 'none';
        primaryIndexInput.value = '0';
        currentFiles = [];
        currentPreviews = [];
        imageInput.value = '';
    });

    if (primarySelector.parentElement) {
        primarySelector.parentElement.appendChild(clearButton);
    }
    document.querySelectorAll('.hover-zoom').forEach(img => {
        img.addEventListener('click', () => {
            img.classList.toggle('enlarged');
        });
    });
});

