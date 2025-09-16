document.addEventListener('click', function (e) {
  const img = e.target.closest('.post-image');
  if (!img) return;

  // Remove enlarged from other images (optional)
  document.querySelectorAll('.post-image.enlarged').forEach(el => {
    if (el !== img) el.classList.remove('enlarged');
  });

  img.classList.toggle('enlarged');
});