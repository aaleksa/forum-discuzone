document.addEventListener('DOMContentLoaded', function () {
    const btn = document.getElementById("notificationsBtn");
    const list = document.getElementById("notificationsList");
    const notificationCount = document.getElementById("notification-count");

    if (!btn || !list || !notificationCount) {
        console.error("Notification elements not found");
        return;
    }

    // Add loading state
    let isLoading = false;
    let notifications = [];

   // Function to update the counter
    function updateUnreadCount() {
        const unreadCount = notifications.filter(n => !n.is_read).length;
        notificationCount.textContent = unreadCount;
        notificationCount.style.display = unreadCount > 0 ? 'block' : 'none';
        
        // Add/remove class for animation
        if (unreadCount > 0) {
            notificationCount.classList.add('has-notifications');
        } else {
            notificationCount.classList.remove('has-notifications');
        }
    }

    // Function to display notifications
    function renderNotifications() {
        list.innerHTML = '';
        
        if (notifications.length === 0) {
            list.innerHTML = '<div class="notification-empty">No notifications</div>';
            return;
        }

        notifications.forEach(n => {
        const el = document.createElement("div");
    const typeText = getNotificationText(n.type);
        const isUnread = !n.is_read;

        el.innerHTML = `
            <div class="notification-content">
            <strong>${n.actor}</strong> ${typeText} your post "${n.post_title}"
            <small>${formatTime(n.created_at)}</small>
            <br>
                <a href="/post_page/${n.post_id}" class="notification-link">Go to this post</a>
            </div>
            ${isUnread ? '<div class="notification-dot"></div>' : ''}
        `;

        el.className = `notification-item ${isUnread ? 'unread' : ''}`;
        el.dataset.id = n.id;

    // Click on the entire notification (except the dot)
        el.addEventListener("click", async (e) => {
        if (e.target.classList.contains('notification-dot')) return;
        if (e.target.classList.contains('notification-link')) return; // links are processed separately below
            await handleNotificationClick(n, el);
        });

   // Click on the link "Go to this post"
    el.querySelector('.notification-link').addEventListener('click', async function(e) {
        e.preventDefault();  // stop the default transition
        await handleNotificationClick(n, el); // mark as read + go
        });

    list.appendChild(el);
});

    }

    // Helper functions
    function getNotificationText(type) {
        const types = {
            'comment': 'commented on',
            'like': 'liked',
            'dislike': 'disliked',
            'mention': 'mentioned you in'
        };
        return types[type] || 'interacted with';
        }

    function formatTime(dateString) {
        const date = new Date(dateString);
        return date.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    }

    async function handleNotificationClick(n, el) {
        if (!n.is_read) {
            try {
                await markAsRead(n.id);
                    n.is_read = true;
                    el.classList.remove("unread");
                    el.querySelector('.notification-dot')?.remove();
                    updateUnreadCount();
            } catch (error) {
                console.error("Error marking notification as read:", error);
            }
        }
        window.location.href = `/post_page/${n.post_id}`;
    }

    async function markAsRead(notificationId) {
        const response = await fetch("/notifications/read", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ notification_id: notificationId })
        });
        return response.ok;
    }

    

   // Main loading function
    async function loadNotifications() {
        if (isLoading) return;
        isLoading = true;

        try {
            btn.classList.add('loading');
            const res = await fetch("/notifications");
            
            if (!res.ok) throw new Error('Failed to load notifications');
            
            notifications = await res.json();
            updateUnreadCount();
            renderNotifications();
            
        } catch (error) {
            console.error("Notifications error:", error);
            list.innerHTML = '<div class="notification-error">Failed to load notifications</div>';
        } finally {
            isLoading = false;
            btn.classList.remove('loading');
        }
    }

   // Click handler for the button
    btn.addEventListener("click", async function() {
        list.classList.toggle("visible");
        
        if (list.classList.contains("visible")) {
            await loadNotifications();
        } else {
            list.innerHTML = "";
        }
    });

   // Load the notification when the page loads
    loadNotifications();

   // Update the notification every 30 seconds
    setInterval(loadNotifications, 30000);
});