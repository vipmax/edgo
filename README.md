# edgo
Yet another console text editor, but with lsp support
![editor](screen.png)

Features:
- `Esc` - exit
- `Control + s` - save file
- `Control + q` - quit
- `Control + d` - duplicate line
- `Control + x` - cut 
- `Control + c` - copy 
- `Control + v` - paste
- `Control + u` - undo
- `Control + r` - redo
- `Shift + arrow` - select text
- `Option + right/left` - smart horizontal movement by words
- `Option + down/up` - smart vertical movement (adds new line)
- `mouse selection`  - select text 
- `mouse double click`  - select word 
- `mouse triple click`  - select line
- `Control + space` - lsp completion
- `Control + h` - lsp hover
- `Control + p` - lsp signature help
- `Control + g / Control + mouse click` - lsp definition


Note: map `Caps lock` to `Control` button, everything will be easier.   


### Installation:

Install Go for mac os:
```
brew install go 
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```
And then:   
```shell
git clone https://github.com/vipmax/edgo && cd edgo
go build 
go install .
```

`edgo` will be available as :
```
edgo [filename]
edgo ~/.zshrc 
```
### Lsp

Following lsp features are supported:
- completion
- hover
- signature help
- definition



Following languages are supported:

`go`
```shell  
go install golang.org/x/tools/gopls@latest
```

`python`
```shell  
pip install -U 'python-lsp-server[all]'
```

`javascript/typescript`
```shell  
npm i -g typescript typescript-language-server
```

`html`
```shell  
npm i -g vscode-langservers-extracted
```

`vue`
```shell  
npm i -g vls
```

`rust`
```shell  
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
rustup component add rust-analyzer
```

`c/c++`
```shell  
glang, go to https://clangd.llvm.org/installation.html
```

`java`
```shell  
# jdtls requires at least Java 17
# check also JAVA_HOME must be set 
brew install jdtls
```

`kotlin`
```shell  
# https://github.com/fwcd/kotlin-language-server
brew install kotlin-language-server
```


### Notes:  
Add alias to  shell environment `nano ~/.zshrc` - `alias edgo="./$pwd./edgo"`


If you are using `tmux` I recommend to add `set-option -g default-terminal "xterm-256color" ` to `~/.tmux.conf`  for shift and option keys. Do not forget apply it as `tmux source-file ~/.tmux.conf`  

If you are using `iterm2` I recommend to use `Natural text editing` preset in `Profiles > Keys > Key Mappings > Presets > Natural text editing > Reset ` 

To get file selection I provided `fzf-edgo.sh` script, add it to your shell  
``` shell
echo 'alias e="sh ~/apps/go/edgo/fzf-edgo.sh"' >> ~/.zshrc
```