" rover.vim - Vim plugin for JDK source search, AST outline, symbol navigation, and markdown tools
if exists("g:roverpluginloaded")
  finish
endif
let g:roverpluginloaded=1

let s:plugindir = expand(expand("<sfile>:p:h:h"))
let s:binary = s:plugindir . "/.bin/rover"

" Commands for binary installation and updates
command! -nargs=* -complete=customlist,s:complete RoverInstall call s:RoverInstallBinaries(-1, <f-args>)
command! -nargs=* -complete=customlist,s:complete RoverUpdate  call s:RoverInstallBinaries(1, <f-args>)

function! s:RoverInstallBinaries(updateBinaries, ...)
  silent !clear
  execute "silent !" . "cd " . s:plugindir . "; " . "go build -o" . " " . s:binary
  echomsg "updated rover.vim plugin binary"
endfunction

" Check binary existence
function! s:ensureBinary()
  if !executable(s:binary)
    call s:RoverInstallBinaries(1)
  endif
  return executable(s:binary)
endfunction

" Right Sidebar Document AST Outline (:RoverOutline)
function! s:RoverOutline()
  if !s:ensureBinary()
    return
  endif

  " Toggle behavior: close if already open
  let l:winNum = bufwinnr('__Rover_Outline__')
  if l:winNum != -1
    execute l:winNum . "wincmd c"
    return
  endif

  let l:targetWin = win_getid()
  let l:file = expand('%:p')
  if empty(l:file)
    echoerr "Please save the file first"
    return
  endif

  let l:cmd = shellescape(s:binary) . " outline " . shellescape(l:file)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:list = json_decode(l:res)
  if empty(l:list)
    echomsg "No outline symbols found in " . expand('%:t')
    return
  endif

  " Open dedicated right sidebar window (width 38)
  execute "botright 38vnew __Rover_Outline__"
  setlocal buftype=nofile bufhidden=wipe noswapfile nowrap nonumber signcolumn=no cursorline
  setlocal filetype=roveroutline

  let b:rover_target_win = l:targetWin
  let b:rover_symbol_lines = []

  let l:lines = []
  for l:item in l:list
    call add(l:lines, l:item.text)
    call add(b:rover_symbol_lines, l:item.lnum)
  endfor

  setlocal modifiable
  call setline(1, l:lines)
  setlocal nomodifiable

  " Mappings inside outline window
  nnoremap <buffer> <silent> <CR> :call <SID>JumpToOutlineSymbol()<CR>
  nnoremap <buffer> <silent> q :close<CR>
endfunction

function! s:JumpToOutlineSymbol()
  let l:lineIdx = line('.') - 1
  if !exists('b:rover_symbol_lines') || l:lineIdx < 0 || l:lineIdx >= len(b:rover_symbol_lines)
    return
  endif
  let l:targetLine = b:rover_symbol_lines[l:lineIdx]
  let l:targetWin = get(b:, 'rover_target_win', 0)
  if l:targetWin != 0 && win_id2win(l:targetWin) != 0
    call win_gotoid(l:targetWin)
    execute l:targetLine
    normal! zt
  endif
endfunction

command! -nargs=0 RoverOutline call s:RoverOutline()
command! -nargs=0 JdkOutline   call s:RoverOutline()

" JDK Source Search (:RoverJdkSearch / :JdkSearch)
function! s:RoverJdkSearch(...)
  if !s:ensureBinary()
    return
  endif

  let l:query = join(a:000, ' ')
  if empty(l:query)
    echoerr "Usage: :RoverJdkSearch <query>"
    return
  endif

  let l:cmd = shellescape(s:binary) . " jdk-search " . shellescape(l:query)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:list = json_decode(l:res)
  if empty(l:list)
    echomsg "No JDK source matches found for: " . l:query
    return
  endif

  call setqflist(l:list, 'r')
  copen
endfunction

command! -nargs=* RoverJdkSearch call s:RoverJdkSearch(<f-args>)
command! -nargs=* JdkSearch      call s:RoverJdkSearch(<f-args>)

" Goto Symbol Declaration (:RoverGoto [symbol])
function! s:RoverGoto(...)
  if !s:ensureBinary()
    return
  endif

  let l:symbol = a:0 > 0 ? join(a:000, ' ') : expand('<cword>')
  if empty(l:symbol)
    echoerr "Usage: :RoverGoto <symbol>"
    return
  endif

  let l:cmd = shellescape(s:binary) . " goto " . shellescape(l:symbol)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:list = json_decode(l:res)
  if empty(l:list)
    echomsg "No declaration found for symbol: " . l:symbol
    return
  endif

  if len(l:list) == 1
    execute "edit +" . l:list[0].lnum . " " . fnameescape(l:list[0].filename)
  else
    call setqflist(l:list, 'r')
    copen
  endif
endfunction

command! -nargs=* RoverGoto call s:RoverGoto(<f-args>)

" JDK Cache Clean (:RoverJdkClean / :JdkClean)
function! s:RoverJdkClean()
  if !s:ensureBinary()
    return
  endif

  let l:cmd = shellescape(s:binary) . " jdk-clean"
  let l:res = system(l:cmd)
  echomsg trim(l:res)
endfunction

command! -nargs=0 RoverJdkClean call s:RoverJdkClean()
command! -nargs=0 JdkClean      call s:RoverJdkClean()

" Markdown Image Paste (:RoverImagePaste / :MarkdownImagePaste)
function! s:RoverImagePaste()
  if !s:ensureBinary()
    return
  endif

  let l:file = expand('%:p')
  if empty(l:file)
    echoerr "Please save the file first"
    return
  endif

  let l:cmd = shellescape(s:binary) . " image-paste " . shellescape(l:file)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:tag = trim(l:res)
  if !empty(l:tag)
    execute "normal A" . l:tag
  endif
endfunction

command! -nargs=0 RoverImagePaste    call s:RoverImagePaste()
command! -nargs=0 MarkdownImagePaste call s:RoverImagePaste()

" Markdown Preview (:RoverPreview / :MarkdownPreview)
function! s:RoverPreview()
  if !s:ensureBinary()
    return
  endif

  let l:file = expand('%:p')
  if empty(l:file)
    echoerr "Please save the file first"
    return
  endif

  let l:cmd = shellescape(s:binary) . " preview " . shellescape(l:file)
  call system(l:cmd)
endfunction

command! -nargs=0 RoverPreview    call s:RoverPreview()
command! -nargs=0 MarkdownPreview call s:RoverPreview()

" Interactive VS Code-like Command Palette (:Rover)
function! s:RoverPalette()
  let l:options = [
        \ "=== 🚀 Rover.vim Command Palette ===",
        \ "1. 🌳 Outline       - Toggle Right Sidebar Document Symbols",
        \ "2. 🎯 Goto Symbol   - Jump to Symbol/Class Declaration",
        \ "3. ☕ JDK Search     - Search JDK Official Source Code",
        \ "4. 🖼️ Paste Image   - Paste Clipboard Image into Markdown",
        \ "5. 🧹 Clean JDK     - Clean Extracted JDK Source Cache",
        \ "6. 🌐 Preview       - Launch Markdown Live Preview",
        \ "Select an action (1-6) or ESC: "
        \ ]
  let l:choice = inputlist(l:options)
  redraw!
  if l:choice == 1
    call s:RoverOutline()
  elseif l:choice == 2
    call s:RoverGoto()
  elseif l:choice == 3
    let l:query = input("JDK Search query: ")
    if !empty(l:query)
      call s:RoverJdkSearch(l:query)
    endif
  elseif l:choice == 4
    call s:RoverImagePaste()
  elseif l:choice == 5
    call s:RoverJdkClean()
  elseif l:choice == 6
    call s:RoverPreview()
  endif
endfunction

command! -nargs=0 Rover call s:RoverPalette()

" Default Keymappings
if get(g:, 'rover_enable_keymaps', 1)
  nnoremap <silent> <leader>rr :Rover<CR>
  nnoremap <silent> <leader>ro :RoverOutline<CR>
  nnoremap <silent> <leader>rg :RoverGoto<CR>
  nnoremap <silent> <leader>rj :RoverJdkSearch<CR>
endif

" Auto-clean JDK cache on Vim exit
function! s:AutoCleanJdkCache()
  if executable(s:binary)
    call system(shellescape(s:binary) . " jdk-clean")
  endif
endfunction

augroup RoverAutoClean
  autocmd!
  autocmd VimLeavePre * call s:AutoCleanJdkCache()
augroup END