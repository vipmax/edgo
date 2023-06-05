# edgo
Yet another console text editor, but with lsp support
![editor](screen.png)
I recommend to map `caps lock` to `control` button for faster writing   

Key bindings:
- `esc` - exit
- `Control + s` - save file
- `Control + q` - quit
- `Control + d` - duplicate line
- `Control + x` - cut 
- `Control + c` - copy 
- `Control + v` - paste
- `shift + arrow` - select text
- `option + arrow` - smart movement
- `control + space` - lsp completion
- `mouse selection`  - select text 
- `mouse double click`  - select word 


Install Go for mac os:
```
brew install go 
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```


Installation:
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

Install lsp for `go`
```shell  
go install golang.org/x/tools/gopls@latest
```

Install lsp for `python`
```shell  
pip install "python-lsp-server[all]"
```

Install lsp for `typescipt`
```shell  
npm install -g typescript typescript-language-server
```

### Notes:  
Add alias to  shell environment `nano ~/.zshrc` - `alias edgo="./$pwd./edgo"`


If you are using `tmux` I recommend to add `set-option -g default-terminal "xterm-256color" ` to `~/.tmux.conf`  for shift and option keys. Do not forget apply it as `tmux source-file ~/.tmux.conf`  

If you are using `iterm2` I recommend to use `Natural text editing` preset in `Profiles > Keys > Key Mappings > Presets > Natural text editing > Reset ` 

To get file selection I provided `fzf-edgo.sh` script, add it to your shell  
``` shell
echo 'alias e="sh ~/apps/go/edgo/fzf-edgo.sh"' >> ~/.zshrc
```