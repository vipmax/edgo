# Edgo
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
- option/alt + arrow - smart movement


Install Go lang for mac os:
```
brew install go # mac
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
```

Installation:
```shell
git clone https://github.com/vipmax/edgo && cd edgo
go build 
go install .
```

Or add alias to  shell environment `nano ~/.zshrc`
```shell
alias edgo="./$pwd./edgo"
```

If you are using `tmux` add `set-option -g default-terminal "xterm-256color" ` to conf for shift and option keys. Do not forget apply it as `tmux source-file ~/.tmux.conf`