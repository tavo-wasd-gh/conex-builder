function isFormValid() {
  const form = document.getElementById('mainForm');
  return form.checkValidity(); // Returns true if the form is valid
}

document.getElementById('openDialogButton').addEventListener('click', () => {
  document.getElementById('overlay').style.display = 'block';
  document.getElementById('dialog').style.display = 'block';
  document.getElementById('openDialogButton').style.display = 'none';
});

document.getElementById('submitDialogButton').addEventListener('click', () => {
  const form = document.getElementById('mainForm');

  ['name', 'email', 'phone'].forEach(id => {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = id;
    input.value = document.getElementById(id).value;
    form.appendChild(input);
  });

  if (isFormValid()) {
    document.getElementById('error-message').style.display = 'none';
    // renderPayPalButton();
  } else {
    document.getElementById('error-message').style.display = 'block';
    // removePayPalButton();
  }

  // const orderIDInput = document.createElement('input');
  // orderIDInput.type = 'hidden';
  // orderIDInput.name = 'paypalOrderID';
  // orderIDInput.value = paypalOrderID;
  // form.appendChild(orderIDInput);
  form.submit();
});

document.getElementById('cancelDialogButton').addEventListener('click', () => {
  document.getElementById('overlay').style.display = 'none';
  document.getElementById('dialog').style.display = 'none';
  document.getElementById('openDialogButton').style.display = 'block';
  document.getElementById('submitDialogButton').style.display = 'inline';
  // removePayPalButton();
});
