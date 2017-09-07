SOURCE = samples/Шпаликов
NAME = md2epub
OUT = out

test: $(NAME)
	./$(NAME) $(SOURCE)
	rm -rf $(OUT)
	unzip -o $(SOURCE).epub -d $(OUT)

$(NAME): build

build: 
	go build

.PHONY: test build