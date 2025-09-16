document.addEventListener('DOMContentLoaded', function() {
  // Отримуємо елементи модального вікна один раз
  const modal = document.getElementById("editCommentModal");
  const idField = document.getElementById("commentId");
  const contentField = document.getElementById("commentContent");
  const editForm = document.getElementById("editCommentForm");

  const closeModalBtn = document.getElementById("closeModalBtn");
    if (closeModalBtn) {
         closeModalBtn.addEventListener("click", closeModal);
    }

  // Функція для відкриття модального вікна
  function openEditModal(id, content) {
    console.log('Opening modal with ID:', id, 'Content:', content);
    
    if (!modal || !idField || !contentField) {
      console.error('Modal elements not found');
      return;
    }
    
    // Декодуємо HTML-сутності
    const decodedContent = content.replace(/&lt;/g, '<')
                                 .replace(/&gt;/g, '>')
                                 .replace(/&amp;/g, '&')
                                 .replace(/&quot;/g, '"')
                                 .replace(/&#39;/g, "'");
    
    idField.value = id;
    contentField.value = decodedContent;
    modal.style.display = "block";
    contentField.focus();
  }

  // Функція для закриття модального вікна
  function closeModal() {
    if (modal) {
      modal.style.display = "none";
      if (idField) idField.value = '';
      if (contentField) contentField.value = '';
    }
  }

  // Додаємо обробники подій для всіх кнопок редагування
  document.querySelectorAll('.edit-btn').forEach(button => {
    button.addEventListener('click', function() {
      const id = this.dataset.id;  // Використовуємо dataset замість getAttribute
      const content = this.dataset.content;
      openEditModal(id, content);
    });
  });

  // Закриття модального вікна при кліку поза ним
  window.addEventListener('click', function(event) {
    if (event.target === modal) {
      closeModal();
    }
  });

  // Закриття модального вікна клавішею Escape
  document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
      closeModal();
    }
  });

  // Обробник відправки форми
  if (editForm) {
    editForm.addEventListener("submit", function(event) {
      event.preventDefault();

      const formData = new FormData(this);
      const submitBtn = this.querySelector('button[type="submit"]');
      const originalText = submitBtn?.textContent || 'Save';

      if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.textContent = 'Saving...';
      }

      fetch("/edit_comment/", {
        method: "POST",
        body: formData,
        headers: {
          'X-Requested-With': 'XMLHttpRequest'
        }
      })
      .then(res => {
        if (res.redirected) {
          window.location.href = res.url;
        } else if (!res.ok) {
          throw new Error(`HTTP error! status: ${res.status}`);
        } else {
          closeModal();
          window.location.reload();
        }
      })
      .catch(error => {
        console.error('Error:', error);
        alert('Failed to save comment. Please try again.');
      })
      .finally(() => {
        if (submitBtn) {
          submitBtn.disabled = false;
          submitBtn.textContent = originalText;
        }
      });
    });
  } else {
    console.warn('Edit form not found');
  }
});