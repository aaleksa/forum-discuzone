document.addEventListener("DOMContentLoaded", function () {
  const toggleBtn = document.getElementById("show-avatar-form");
  const form = document.getElementById("avatar-form");

  if (toggleBtn && form) {
    toggleBtn.addEventListener("click", function () {
      form.style.display = form.style.display === "none" ? "flex" : "none";
    });
  } else {
    console.error("Avatar toggle button or form not found in DOM");
  }
});

document.addEventListener('DOMContentLoaded', function() {
    // Tab switching functionality
    const tabButtons = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');
    
    // Check if there are tabs on the page
    if (tabButtons.length === 0 || tabContents.length === 0) {
        console.warn('No tabs found on the page');
        return;
    }
    
    // Function for switching tabs
    function switchTab(tabName) {
       // Hide all tabs
        tabContents.forEach(content => {
            content.classList.remove('active');
        });
        
       // Remove the active class from all buttons
        tabButtons.forEach(btn => {
            btn.classList.remove('active');
        });
        
       // Show the selected tab
        const activeTab = document.getElementById(`${tabName}-tab`);
        if (activeTab) {
            activeTab.classList.add('active');
        } else {
            console.error(`Tab with id ${tabName}-tab not found`);
        }
        
       // Make the button active
        const activeBtn = document.querySelector(`.tab-btn[data-tab="${tabName}"]`);
        if (activeBtn) {
            activeBtn.classList.add('active');
        }
    }
    
   // Click handler for tab buttons
    tabButtons.forEach(button => {
        button.addEventListener('click', function() {
            const tabName = this.getAttribute('data-tab');
            switchTab(tabName);
        });
    });
    
    // Activate the first tab by default
    const defaultTab = tabButtons[0].getAttribute('data-tab');
    switchTab(defaultTab);
});