# rover.vim

*A powerful Vim plugin written in Go for code navigation, JDK source code browsing, AST class outlines, and Markdown editing.*

## Features

- 🌳 **AST Class & Document Outline** (`:RoverOutline` / `:JdkOutline`): Instant AST-based document symbol outline for **Go** (via Go stdlib `go/ast`) and **Java** (via Tree-Sitter). Displays classes, interfaces, structs, methods, and fields in Vim's Location List (`:lopen`).
- 🎯 **Goto Symbol & Declaration** (`:RoverGoto [symbol]`): Jump to target class, struct, variable, or method declaration in workspace and JDK source code.
- ☕ **JDK Source Browsing & Search** (`:JdkSearch <query>`): High-performance JDK source code search powered by [`sniphunt`](https://github.com/qtopie/sniphunt). Automatically locates `$JAVA_HOME` or system JDK, extracts `src.zip`, and populates Vim's Quickfix list to jump directly to target Java source files and line locations.
- 🧹 **Automatic JDK Cache Cleanup** (`:JdkClean`): Automatically cleans up extracted JDK source cache when closing Java source buffers or exiting Vim.
- 🖼️ **Markdown Asset & Clipboard Integration** (`:MarkdownImagePaste`): Paste images directly from system clipboard into Markdown documents and generate assets automatically.
- 🌐 **Markdown Live Preview** (In progress): High-performance Markdown live preview server written in Go.

## Installation

Add `rover.vim` to your plugin manager. Example using `vim-plug`:

```vim
Plug 'qtopie/rover.vim', { 'do': ':RoverUpdate' }
```

After adding the line to your `.vimrc`, run `:PlugInstall` in Vim.

## Usage

### Document Outline (Go & Java)

Display structured class/struct outline tree for current file:

```vim
:RoverOutline
```

- Displays outline in Vim Location List (`:lopen`). Press `<CR>` to jump to any symbol definition.

### Goto Symbol Declaration

Jump to declaration of a class, struct, method, or variable:

```vim
:RoverGoto PriorityQueue
```
*(Or run `:RoverGoto` with cursor on any symbol).*

### JDK Source Code Browsing

Search for Java standard library classes, interfaces, methods, or symbols:

```vim
:JdkSearch PriorityQueue
```
or
```vim
:JdkSearch class PriorityQueue
```

- Results open automatically in Vim's Quickfix list (`:copen`).
- Press `<CR>` on any result line to open the JDK source code file at that exact line.
- To manually clean the JDK source cache:
  ```vim
  :JdkClean
  ```

### Markdown Tools

Paste an image from your system clipboard into a Markdown document:

```vim
:MarkdownImagePaste
```

## License

MIT
