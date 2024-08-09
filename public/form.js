// Handle form submission after filling out the dialog
document.getElementById('openDialogButton').addEventListener('click', () => {
  document.getElementById('dialog').style.display = 'block';
});

document.getElementById('submitDialogButton').addEventListener('click', () => {
  const form = document.getElementById('mainForm');
  const name = document.getElementById('name').value;
  const email = document.getElementById('email').value;
  const phone = document.getElementById('phone').value;

  // Append additional fields to the form
  const nameInput = document.createElement('input');
  nameInput.type = 'hidden';
  nameInput.name = 'name';
  nameInput.value = name;
  form.appendChild(nameInput);

  const emailInput = document.createElement('input');
  emailInput.type = 'hidden';
  emailInput.name = 'email';
  emailInput.value = email;
  form.appendChild(emailInput);

  const phoneInput = document.createElement('input');
  phoneInput.type = 'hidden';
  phoneInput.name = 'phone';
  phoneInput.value = phone;
  form.appendChild(phoneInput);

  // Submit
  form.submit();
});

// Cancel button
document.getElementById('cancelDialogButton').addEventListener('click', () => {
  // Clear values
  document.getElementById('name').value = '';
  document.getElementById('email').value = '';
  document.getElementById('phone').value = '';
  // Hide dialog
  document.getElementById('dialog').style.display = 'none';
});
