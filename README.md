# rover.vim

*A powerful Vim plugin written in Go for code navigation, JDK source code browsing, right-sidebar AST class outlines, and VS Code-like Command Palette.*

## Features

- 🚀 **VS Code-style Command Palette** (`:Rover`): Interactive command picker so you don't need to memorize individual commands! Simply run `:Rover` or press `<leader>rr`.
- 🌳 **Right Sidebar Document Outline** (`:RoverOutline`): Instant AST-based document symbol sidebar for **Go** (via Go stdlib `go/ast`) and **Java** (via Tree-Sitter). Opens a dedicated right sidebar window showing classes, interfaces, structs, methods, and fields.
- 🎯 **Goto Symbol & Declaration** (`:RoverGoto [symbol]`): Jump to target class, struct, variable, or method declaration in workspace and JDK source code.
- ☕ **JDK Source Browsing & Search** (`:RoverJdkSearch <query>`): High-performance JDK source code search powered by [`sniphunt`](https://github.com/qtopie/sniphunt). Automatically locates `$JAVA_HOME`, extracts `src.zip`, and populates Quickfix list to jump directly to target Java source files.
- 🧹 **Automatic JDK Cache Cleanup** (`:RoverJdkClean`): Automatically cleans up extracted JDK source cache when closing Java source buffers or exiting Vim.
- 🖼️ **Markdown Asset & Clipboard Integration** (`:RoverImagePaste`): Paste images directly from system clipboard into Markdown documents and generate assets automatically.
- 🌐 **Markdown Live Preview** (`:RoverPreview`): High-performance Markdown live preview server written in Go.

## Installation

Add `rover.vim` to your plugin manager. Example using `vim-plug`:

```vim
Plug 'qtopie/rover.vim', { 'do': ':RoverUpdate' }
```

After adding the line to your `.vimrc`, run `:PlugInstall` in Vim.

## Key Mappings

`rover.vim` enables standard shortcuts by default:

| Shortcut | Command | Action |
| :--- | :--- | :--- |
| **`<leader>rr`** | `:Rover` | Open VS Code-style Command Palette |
| **`<leader>ro`** | `:RoverOutline` | Toggle Right Sidebar AST Outline |
| **`<leader>rg`** | `:RoverGoto` | Jump to Symbol / Declaration |
| **`<leader>rj`** | `:RoverJdkSearch` | Search JDK Source Code |

*(Disable default mappings in `.vimrc` via `let g:rover_enable_keymaps = 0`).*

## Commands

All commands are unified under the **`Rover`** prefix:

| Command | Description |
| :--- | :--- |
| **`:Rover`** | Open interactive VS Code-style Command Palette |
| **`:RoverOutline`** | Toggle Right Sidebar AST Document Symbol Outline |
| **`:RoverGoto [symbol]`** | Jump to Class / Method / Variable declaration |
| **`:RoverJdkSearch <query>`** | Search JDK official source code |
| **`:RoverJdkClean`** | Clean extracted JDK source cache |
| **`:RoverImagePaste`** | Paste image from system clipboard into Markdown |
| **`:RoverPreview`** | Launch Markdown live preview |

## License

MIT
