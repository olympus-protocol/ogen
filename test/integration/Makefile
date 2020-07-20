test_sync:
	go test -tags sync_test -c sync_test.go
	./sync.test
	make clean
	
clean:
	rm -rf data_*	
	rm -rf *.test