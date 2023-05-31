# edgo
Yet another console text editor
![editor](screen.png)
I recommend to map `caps lock` to `control` button for faster writing   

Key bindings:
- Control + s - save file
- Control + q - quit
- Control + d - duplicate line
- Control + x - cut 
- Control + c - copy 
- Control + v - paste
- shift + arrow - select text
- option/alt + arrow - smart movenment


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


Alternatives:  
Add alias to  shell environment `nano ~/.zshrc` - `alias edgo="./$pwd./edgo"`


If you are using `tmux` i recommend to add `set-option -g default-terminal "xterm-256color" ` to `~/.tmux.conf`  for shift and option keys. Do not forget apply it as `tmux source-file ~/.tmux.conf`

If you are using `iterm2` I recommend to use `Natural text editing` preset in `Profiles > Keys > Key Mappings > Presets > Natural text editing > Reset ` 