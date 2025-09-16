document.addEventListener('DOMContentLoaded', function () {

    function markAllAsRead() {
        console.log('Mark all as read called');

        fetch('/notifications/read-all', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Requested-With': 'XMLHttpRequest'
            },
            credentials: 'same-origin'
        })
        .then(response => {
            if (response.ok) {
                updateUIForReadNotifications();
                const badge = document.querySelector('.notif-badge');
                if (badge) badge.remove();
            } else {
                console.error('Failed to mark all notifications as read');
            }
        })
        .catch(error => {
            console.error('Error marking all as read:', error);
        });
    }

    function updateUIForReadNotifications() {
        document.querySelectorAll('.notif-item--unread').forEach(item => {
            item.classList.remove('notif-item--unread');
            const dot = item.querySelector('.notif-dot');
            if (dot) dot.remove();
        });
    }

    // Обробник кнопки "Mark all as read"
    const markAllBtn = document.getElementById('mark-all-read-btn');
    if (markAllBtn) {
        markAllBtn.addEventListener('click', function() {
            markAllAsRead();
        });
    }

    // **Видалена логіка закриття/відкриття випадашки**

    // Mark individual notification as read when clicked
    document.querySelectorAll('.notif-item').forEach(item => {
        item.addEventListener('click', function() {
            const notifId = this.getAttribute('data-id');
            const isUnread = this.classList.contains('notif-item--unread');
            
            if (isUnread && notifId) {
                fetch('/notifications/read/' + notifId, {
                    method: 'POST',
                    headers: { 
                        'Content-Type': 'application/json',
                        'X-Requested-With': 'XMLHttpRequest'
                    },
                    credentials: 'same-origin'
                })
                .then(response => {
                    if (response.ok) {
                        this.classList.remove('notif-item--unread');
                        const dot = this.querySelector('.notif-dot');
                        if (dot) dot.remove();
                        
                        const badge = document.querySelector('.notif-badge');
                        if (badge) {
                            const currentCount = parseInt(badge.textContent);
                            if (currentCount <= 1) {
                                badge.remove();
                            } else {
                                badge.textContent = currentCount - 1;
                            }
                        }
                    }
                })
                .catch(error => {
                    console.error('Error marking notification as read:', error);
                });
            }
        });
    });

    function formatTime(dateString) {
        const date = new Date(dateString);
        return date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    }
    document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll('.notif-time').forEach(function (el) {
      const rawDate = el.dataset.time;
      el.textContent = formatTime(rawDate);
    });
  });

    window.markAllAsRead = markAllAsRead;

});

