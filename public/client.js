document.addEventListener("DOMContentLoaded", function() {
  const dialog = document.getElementById("dialog");
  const overlay = document.getElementById("overlay");
  const menu = document.getElementById("floatingButtons");

  function openDialog() {
    dialog.style.display = "block";
    overlay.style.display = "block";
    menu.style.display = "none";
  }

  function closeDialog() {
    dialog.style.display = "none";
    overlay.style.display = "none";
    menu.style.display = "block";
  }

  document.getElementById("openDialogButton").addEventListener("click", openDialog);
  document.getElementById("cancelDialogButton").addEventListener("click", closeDialog);
});
