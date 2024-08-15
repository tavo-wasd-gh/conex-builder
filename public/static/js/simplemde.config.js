var simplemde = new SimpleMDE({
  element: document.getElementById("editor"),
  autosave: {
    enabled: true,
    uniqueId: "main-editor",
    delay: 1000,
  },
  toolbar: ["preview", "|", "heading", "bold", "italic", "unordered-list", "ordered-list", "|", "link", "image", "table"],
  spellChecker: false,
  status: false,
  placeholder: "Contruye tu página aquí utilizando la barra de herramientas de arriba.\n\nRecuerde editar también\n[Nombre Ejemplo] y [Slogan]."
});
