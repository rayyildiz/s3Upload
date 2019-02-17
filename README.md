# S3Uploader (Deprecated) Use [Aws CLI](https://docs.aws.amazon.com/cli/latest/reference/s3/index.html#single-local-file-and-s3-object-operations)

[![Build Status](https://travis-ci.org/rayyildiz/s3Upload.svg?branch=master)](https://travis-ci.org/rayyildiz/s3Upload)

CLI for uploading files to AWS S3

## Usage

Copy all ```*.txt``` files under the ```/home/rayyildiz/documents``` folder to your S3 bucket ```your-bucket-name```.

```bash
s3Upload -s3-bucket=your-bucket-name -directory-root=/home/rayyildiz/documents -file-extension=txt
``` 
