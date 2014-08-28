SOURCE = test
NAME = md2epub
OUT = out

test: $(NAME) $(OUT)
	./$(NAME) $(SOURCE)
	unzip -o $(SOURCE).epub -d $(OUT)

$(NAME): build

build: 
	go build

$(OUT):
	rm -rf $(OUT)

.PHONY: test build