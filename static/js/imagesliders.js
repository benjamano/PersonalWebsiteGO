const containers = document.querySelectorAll('.beforeAfterContainer');

containers.forEach((container, index) => {
    const slider = container.querySelector('.slider');
    const afterImage = container.querySelector('.after');
    let isDragging = false;
    let animationId = null;

    function updateSlider(x) {
        if (animationId) {
            cancelAnimationFrame(animationId);
        }
        
        animationId = requestAnimationFrame(() => {
            const containerRect = container.getBoundingClientRect();
            const containerWidth = containerRect.width;
            const relativeX = x - containerRect.left;
            const percentage = Math.max(0, Math.min(100, (relativeX / containerWidth) * 100));
            
            slider.style.left = percentage + '%';
            afterImage.style.clipPath = `polygon(0 0, ${percentage}% 0, ${percentage}% 100%, 0 100%)`;
        });
    }

    container.addEventListener('mousedown', (e) => {
        isDragging = true;
        container.style.cursor = 'col-resize';
        updateSlider(e.clientX);
        e.preventDefault();
    });

    document.addEventListener('mousemove', (e) => {
        if (isDragging) {
            updateSlider(e.clientX);
        }
    });

    document.addEventListener('mouseup', () => {
        isDragging = false;
        container.style.cursor = 'col-resize';
    });

    container.addEventListener('touchstart', (e) => {
        isDragging = true;
        updateSlider(e.touches[0].clientX);
        e.preventDefault();
    }, { passive: false });

    document.addEventListener('touchmove', (e) => {
        if (isDragging) {
            updateSlider(e.touches[0].clientX);
            e.preventDefault();
        }
    }, { passive: false });

    document.addEventListener('touchend', () => {
        isDragging = false;
    });

    container.addEventListener('dragstart', (e) => {
        e.preventDefault();
    });
});