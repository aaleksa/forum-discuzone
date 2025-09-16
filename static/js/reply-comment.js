document.addEventListener('DOMContentLoaded', function () {
   // Processing form submission via AJAX
    document.querySelectorAll('.reply-form').forEach(form => {
        form.addEventListener('submit', function (e) {
            e.preventDefault();

            const submitBtn = form.querySelector('.submit-reply-btn');
            const originalText = submitBtn.innerHTML;
            submitBtn.disabled = true;
            submitBtn.innerHTML = '<span class="spinner">Sending...</span>';

            const formData = new FormData(this);

            fetch('/notifications/add_reply', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: new URLSearchParams({
                    parent_comment_id: formData.get('parent_comment_id'),
                    post_id: formData.get('post_id'),
                    reply_content: formData.get('reply_content')
                })
            })
                .then(response => {
                    const contentType = response.headers.get('content-type');
                    if (contentType && contentType.includes('application/json')) {
                        return response.json();
                    } else {
                        return response.text().then(text => {
                            throw new Error(`Server returned non-JSON: ${text}`);
                        });
                    }
                })
                .then(data => {
                    if (data.success) {
                        form.style.display = 'none';
                        const replyBtn = form.previousElementSibling;
                        replyBtn.textContent = '↩️ Reply';

                        if (typeof addReplyToDOM === 'function') {
                            addReplyToDOM(data.reply);
                        } else {
                            location.reload();
                        }
                    } else {
                        alert('Error: ' + data.message);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('An error occurred while submitting the reply');
                })
                .finally(() => {
                    submitBtn.disabled = false;
                    submitBtn.innerHTML = originalText;
                });
        });
    });
});
