clear

# hide cursor


while true; do
    tput civis
    # Use fd to generate a list of files and directories, excluding .git and __pycache__
    fd_output=$(fd --type f --follow --exclude .git --exclude __pycache__ --exclude node_modules)
    tput civis
    # Pipe the output of fd into fzf
#    result=$(echo "$fd_output" | fzf --reverse --preview 'bat --style numbers,changes --color=always {} --line-range :500')
    result=$(echo "$fd_output" | fzf --reverse)
    # Exit code of fzf when Esc is pressed is 130
    if [ $? -eq 130 ]; then
        break
    fi

    tput civis
    # Run the edgo command with the result
    edgo $result
done


# alias e="sh ~/apps/go/edgo/fzf-edgo.sh"

