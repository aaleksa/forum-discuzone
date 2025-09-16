document.addEventListener("DOMContentLoaded", function () {
    // 🔁 Forgot password form
    const resetForm = document.getElementById("resetForm");
    if (resetForm) {
        resetForm.addEventListener("submit", async function (e) {
            e.preventDefault();
            const formData = new FormData(resetForm);

            const res = await fetch("/forgot-password-submit", {
                method: "POST",
                body: formData,
            });

            if (res.ok) {
                const modal = document.getElementById("successModal");
                if (modal) modal.style.display = "flex";
                resetForm.reset();
            } else {
                alert("Error sending reset link");
            }
        });
    }

    // 🔁 Reset password form
    const resetPasswordForm = document.getElementById("resetPasswordForm");
    if (resetPasswordForm) {
        resetPasswordForm.addEventListener("submit", async function (e) {
            e.preventDefault();

            const password = document.getElementById("password")?.value;
            const confirm = document.getElementById("confirmPassword")?.value;

            if (password !== confirm) {
                alert("Passwords do not match.");
                return;
            }

            const formData = new FormData(resetPasswordForm);
            const res = await fetch("/reset-password-submit", {
                method: "POST",
                body: formData,
            });

            if (res.ok) {
                const modal = document.getElementById("successModal");
                if (modal) modal.style.display = "flex";
                resetPasswordForm.reset();
            } else {
                alert("Error resetting password.");
            }
        });
    }
    
    document.querySelectorAll(".commentForm").forEach((form) => {
        const textarea = form.querySelector(".commentContent");
        const modal = form.querySelector(".commentModal");
        const closeBtn = form.querySelector(".closeCommentModalBtn");

        if (!textarea || !modal) return;

        // Обробка сабміту
        form.addEventListener("submit", (e) => {
            const content = textarea.value.trim();
            console.log("➡️ Submitted comment:", content);

            if (!content) {
                e.preventDefault();
                modal.style.display = "flex";
            }
        });

        // Закриття модалки
        closeBtn?.addEventListener("click", () => {
            modal.style.display = "none";
        });
    });

});

// 🔁 Close success modal and redirect
function closeModal() {
    const modal = document.getElementById("successModal");
    if (modal) modal.style.display = "none";
    window.location.href = "/login";
}

