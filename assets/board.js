function updateAllCounts() {
    const cards = ['card1', 'card2', 'card3'];
    cards.forEach(cardType => {
        const items = document.querySelectorAll(`#${cardType} .drag-item`);
        const countSpan = document.getElementById(`${cardType}-count`);
        countSpan.textContent = items.length;
    });
}

const items = document.querySelectorAll('.drag-item');
let draggedItem = null;

items.forEach(item => {
    item.addEventListener('dragstart', function (e) {
        draggedItem = item;
        setTimeout(() => {
            item.classList.add('dragging');
        }, 0);
    });

    item.addEventListener('dragend', function (e) {
        setTimeout(() => {
            draggedItem = null;
            item.classList.remove('dragging');
            updateAllCounts();
        }, 0);
    });
});

const cards = document.querySelectorAll('.drag-card');

cards.forEach(card => {
    card.addEventListener('dragover', function (e) {
        e.preventDefault();
    });

    card.addEventListener('dragenter', function (e) {
        e.preventDefault();
    });

    card.addEventListener('drop', function (e) {
        if (draggedItem) {
            card.appendChild(draggedItem);

            const itemId = draggedItem.getAttribute('data-id');
            const newStatus = card.id;

            fetch('/update-issue-status', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ id: itemId, status: newStatus }),
            })
            .then(response => response.json())
            .then(data => {
                console.log('Başarılı:', data);
                updateAllCounts(); // Burada da güncelleyin
            })
            .catch((error) => {
                console.error('Hata:', error);
            });

            updateAllCounts(); // Hemen güncelleyin (ağ isteğini beklemeden)
        }
    });
});

// Sayfa yüklendiğinde sayımı başlatın
document.addEventListener('DOMContentLoaded', updateAllCounts);