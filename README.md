# rover.vim

*A powerful Vim plugin written in Go for code navigation, JDK source code browsing, and Markdown editing.*

## Features

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

### JDK Source Code Browsing

Search for Java standard library classes, interfaces, methods, or symbols:

```vim
:JdkSearch PriorityQueue
```
or
```vim
:JdkSearch class PriorityQueue
```

- Results will open automatically in Vim's Quickfix list (`:copen`).
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
