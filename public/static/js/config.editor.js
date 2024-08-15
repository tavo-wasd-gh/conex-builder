var editor = new EditorJS({
  readOnly: false,
  holder: 'editorjs',

  inlineToolbar: ['link', 'marker', 'bold', 'italic'],
  inlineToolbar: true,
  tools: {
    /**
     * Each Tool is a Plugin. Pass them via 'class' option with necessary settings {@link docs/tools.md}
     */
    header: {
      class: Header,
      inlineToolbar: ['marker', 'link'],
      config: {
        placeholder: 'Header'
      },
      shortcut: 'CMD+SHIFT+H'
    },

    image: SimpleImage,

    list: {
      class: List,
      inlineToolbar: true,
      shortcut: 'CMD+SHIFT+L'
    },

    quote: {
      class: Quote,
      inlineToolbar: true,
      config: {
        quotePlaceholder: 'Enter a quote',
        captionPlaceholder: 'Quote\'s author',
      },
      shortcut: 'CMD+SHIFT+O'
    },

    linkTool: LinkTool,

    embed: Embed,

    table: {
      class: Table,
      inlineToolbar: true,
      shortcut: 'CMD+ALT+T'
    },

  },

  /**
   * This Tool will be used as default
   */
  // defaultBlock: 'paragraph',

  data: {
    blocks: [
      {
        type: "header",
        data: {
          text: "Acerca de [Empresa]",
          level: 1
        }
      },
      {
        type : 'paragraph',
        data : {
          text : 'En [Nombre de Tu Empresa], nos dedicamos a ofrecer [tu servicio/producto] de la más alta calidad con un servicio al cliente excepcional. Nuestro equipo de expertos se asegura de que cada aspecto de tu experiencia sea manejado con profesionalismo y cuidado.'
        }
      },
      {
        type : 'list',
        data : {
          items : [
            'It is a block-styled editor',
            'It returns clean data output in JSON',
            'Designed to be extendable and pluggable with a simple API',
          ],
          style: 'unordered'
        }
      },
      {
        type: 'table',
        data: {
          content: [
            ['Header 1', 'Header 2', 'Header 3'],
            ['Row 1, Cell 1', 'Row 1, Cell 2', 'Row 1, Cell 3'],
            ['Row 2, Cell 1', 'Row 2, Cell 2', 'Row 2, Cell 3']
          ]
        }
      },
    ]
  },
  i18n: {
    messages: {
      ui: {
        "blockTunes": {
          "toggler": {
            "Click to tune": "Modificar",
            "or drag to move": "or drag to move"
          },
        },
        "inlineToolbar": {
          "converter": {
            "Convert to": "Convertir a"
          }
        },
        "toolbar": {
          "toolbox": {
            "Add": "Insertar"
          }
        }
      },

      /**
       * Section for translation Tool Names: both block and inline tools
       */
      toolNames: {
        "Text": "Texto",
        "Heading": "Título",
        "List": "Lista",
        "Warning": "Advertencia",
        "Quote": "Quote",
        "Table": "Tabla",
        "Link": "Link",
        "Bold": "Negrita",
        "Italic": "Itálicas",
        "InlineCode": "InlineCode",
      },

      /**
       * Section for passing translations to the external tools classes
       */
      tools: {
        /**
         * Each subsection is the i18n dictionary that will be passed to the corresponded plugin
         * The name of a plugin should be equal the name you specify in the 'tool' section for that plugin
         */
        "warning": { // <-- 'Warning' tool will accept this dictionary section
          "Title": "Título",
          "Message": "Mensaje",
        },

        /**
         * Link is the internal Inline Tool
         */
        "link": {
          "Add a link": "Agregar link"
        },
        /**
         * The "stub" is an internal block tool, used to fit blocks that does not have the corresponded plugin
         */
        "stub": {
          'The block can not be displayed correctly.': 'No se puede visualizar este bloque'
        }
      },

      /**
       * Section allows to translate Block Tunes
       */
      blockTunes: {
        /**
         * Each subsection is the i18n dictionary that will be passed to the corresponded Block Tune plugin
         * The name of a plugin should be equal the name you specify in the 'tunes' section for that plugin
         *
         * Also, there are few internal block tunes: "delete", "moveUp" and "moveDown"
         */
        "delete": {
          "Delete": "Quitar bloque"
        },
        "moveUp": {
          "Move up": "Mover arriba"
        },
        "moveDown": {
          "Move down": "Mover abajo"
        }
      },
    }
  },
  onReady: function(){
    saveButton.click();
  },
  onChange: function(api, event) {
    console.log('something changed', event);
  }

});

/**
 * Saving button
 */
const saveButton = document.getElementById('saveButton');

/**
 * Toggle read-only button
 */
const toggleReadOnlyButton = document.getElementById('toggleReadOnlyButton');
const readOnlyIndicator = document.getElementById('readonly-state');

/**
 * Saving example
 */
saveButton.addEventListener('click', function () {
  editor.save()
    .then((savedData) => {
      cPreview.show(savedData, document.getElementById("output"));
    })
    .catch((error) => {
      console.error('Saving error', error);
    });
});

/**
 * Toggle read-only example
 */
toggleReadOnlyButton.addEventListener('click', async () => {
  const readOnlyState = await editor.readOnly.toggle();

  readOnlyIndicator.textContent = readOnlyState ? 'On' : 'Off';
});
