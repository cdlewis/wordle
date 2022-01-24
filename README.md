# Wordle Solver

Build and find most optimal starting word:

```
go build
./wordle
```

You can also run it with your own constraints to find the next most optimal word, e.g

```
./wordle \
    -withoutLetters reist \
    -withLettersAtPosition o=2 \
    -withLettersNotAtPosition l=1
```
