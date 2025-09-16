
document.addEventListener('DOMContentLoaded', function() {
    const checkboxContainer = document.getElementById('checkboxContainer');
    const selectedCount = document.getElementById('selectedCount');
    const selectedTags = document.getElementById('selectedTags');
    const clearBtn = document.getElementById('clearSelection');
    const maxSelected = 3;
    
    let selectedValues = new Set();
    
    // Initialization from default values
    document.querySelectorAll('.category-checkbox:checked').forEach(checkbox => {
        selectedValues.add(checkbox.value);
    });
    
    updateDisplay();
    
    // Checkbox click handler
    checkboxContainer.addEventListener('change', function(e) {
        if (e.target.classList.contains('category-checkbox')) {
            const checkbox = e.target;
            const value = checkbox.value;
            
            if (checkbox.checked) {
                if (selectedValues.size < maxSelected) {
                    selectedValues.add(value);
                } else {
                    checkbox.checked = false;
                    return;
                }
            } else {
                selectedValues.delete(value);
            }
            
            updateDisplay();
        }
    });
    
    // Checkbox click handler
    clearBtn.addEventListener('click', function() {
        selectedValues.clear();
        document.querySelectorAll('.category-checkbox').forEach(cb => cb.checked = false);
        updateDisplay();
    });
    
    // Refresh display
    function updateDisplay() {
        // Update counter
        selectedCount.textContent = selectedValues.size;
        
        //Update selected tags
        selectedTags.innerHTML = '';
        selectedValues.forEach(value => {
            const checkbox = document.querySelector(`.category-checkbox[value="${value}"]`);
            if (checkbox) {
                const tag = document.createElement('div');
                tag.className = 'selected-tag';
                tag.innerHTML = `
                    <button type="button" class="remove-tag" data-value="${value}"> ${checkbox.dataset.name} ×</button>
                `;
                selectedTags.appendChild(tag);
            }
        });
        
        //Add delete handlers
        document.querySelectorAll('.remove-tag').forEach(btn => {
            btn.addEventListener('click', function() {
                const value = this.getAttribute('data-value');
                selectedValues.delete(value);
                document.querySelector(`.category-checkbox[value="${value}"]`).checked = false;
                updateDisplay();
            });
        });
        
        //Blocking checkboxes when the limit is reached
        document.querySelectorAll('.category-checkbox').forEach(checkbox => {
            checkbox.disabled = (selectedValues.size >= maxSelected) && !checkbox.checked;
        });
    }
});