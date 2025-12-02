# Tree-sitter grammar for Alna

This is a Tree-sitter parser for the Alna programming language.

## Setup

### Prerequisites

1. Install Node.js and npm
2. Install tree-sitter CLI:
   ```bash
   npm install -g tree-sitter-cli
   ```

### Build the parser

```bash
cd tree-sitter-alna
npm install
tree-sitter generate
```

### Test the parser

Create test files in `test/corpus/`:

```bash
tree-sitter test
```

### Try it out

```bash
tree-sitter parse ../examples/conditionals.alna
```

## Neovim Integration

### Option 1: Manual Setup (Local Development)

1. Build the parser:
   ```bash
   cd tree-sitter-alna
   tree-sitter generate
   ```

2. Create parser directory in Neovim:
   ```bash
   mkdir -p ~/.local/share/nvim/site/parser
   ```

3. Compile and copy the parser:
   ```bash
   cc -o ~/.local/share/nvim/site/parser/alna.so \
      -shared src/parser.c \
      -I./src -Os
   ```

4. Add to your Neovim config (`~/.config/nvim/init.lua` or similar):
   ```lua
   -- Register the Alna filetype
   vim.filetype.add({
     extension = {
       alna = 'alna',
     }
   })

   -- Configure tree-sitter for Alna
   local parser_config = require('nvim-treesitter.parsers').get_parser_configs()
   parser_config.alna = {
     install_info = {
       url = "~/path/to/alna-lang/tree-sitter-alna",
       files = {"src/parser.c"},
       branch = "main",
     },
     filetype = "alna",
   }

   -- Set up queries path
   vim.treesitter.query.set("alna", "highlights", [[
     ; Your highlights.scm content here
   ]])
   ```

5. Copy highlight queries:
   ```bash
   mkdir -p ~/.config/nvim/queries/alna
   cp queries/highlights.scm ~/.config/nvim/queries/alna/
   ```

### Option 2: Using nvim-treesitter (Recommended)

After building the parser, you can use nvim-treesitter's `:TSInstall` with a local path.

Add to your nvim-treesitter config:
```lua
require('nvim-treesitter.configs').setup({
  parser_install_dir = "~/.local/share/nvim/site",
  ensure_installed = { "alna" },
  highlight = {
    enable = true,
  },
})
```

## Development

- Edit `grammar.js` to modify the grammar
- Run `tree-sitter generate` to regenerate the parser
- Run `tree-sitter test` to run tests
- Run `tree-sitter parse <file>` to test parsing individual files

## Testing

Create test files in `test/corpus/` directory with the format:

```
==================
Test name
==================

int x = 10

---

(source_file
  (variable_declaration
    type: (type)
    name: (identifier)
    value: (number)))
```
