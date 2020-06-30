# Usage: sh file.sh output_filename.ext
# reads from stdin and writes out to a file named on the command line.
while read line; do
  echo "$line" >> $1
done < /dev/stdin
