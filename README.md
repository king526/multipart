# multipart

this repository is refine official multipart lib.

1. Allow specified temp multipart file save path,which may use large space.
2. can move temp file to dest instead of copy write.


## FormBody 
	when use Writer make a multipart form request with large file,the form bytes  
    need to write to file or memory.

	FormBody add CreateFromReader,CreateFromByPath for no read file to form bytes,
	it read the Reader(file) at post request.