# String-a-ling
Stringaling is a streaming string replacement tool for handling string replacement in extremely large files 
in an environment with limited resources.

Its going to take a long time for large files (about 40 minutes with verbosity for a 200mb xml file)
but it will get the job done.

## Expectations
It's important to set expectations for this program/library.
This does NOT support regex at all, and should only be used for known token replacements.

### Usage
The basic usage for all commands is 

```bash
$ stringaling COMMAND [-v] ...
```

The very first argument must always be the command you wish to run.

All commands support the -v (verbose) option which outputs debug information to STDOUT.

#### Replace All
 
This command replaces all text between two tokens (including the tokens) with another token.
The syntax of this command is 

```bash
$ stringaling replace-all|ra [-v] -i INPUT_FILE -o OUTPUT_FILE -s START_TOKEN -e END_TOKEN [-w TOKEN] [-t THREADS]
``` 

The command can either be `replace-all` or `ra` for short.

##### Minimum Requirements 
ReplaceAll's requirements scale with the size and complexity of the file which the replacement is happening on.
ReplaceAll keeps an in-memory cache of potentially skipped characters, as it may need to write these characters to the output
if a replacement never finishes, this is especially an issue when using threading.

That said, the memory needed should never exceed the unit of work.
So a 200mb file safely needs 200mb of memory to process, however on average, this number can be divided by the number of threads used.
So a 200mb file on 5 threads likely won't use more than 40mb of memory, but in the worst case, all threads could use 40mb at a time, or 200mb.

In a future release, an option may be given to use a file-based cache instead of memory, which may prove useful for gigantic files (in the GB+ range)

##### Arguments
* -i INPUT_FILE 
  * The input file to perform the action on
* -o OUTPUT_FILE
  * The output file to write the action to
* -s START_TOKEN
  * The token to mark the beginning of a replacement
* -e END_TOKEN
  * The token to mark the end of replacement
* -w TOKEN
  * The token to use as a replacement, default is emptystring 
* -t THREADS
  * The number of threads to use, defaults to 1, for optimum performance, set this to the number of cores available

##### Example
Given input file `results.xml`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<root>
    <test name="my test">
        <phi>
            <name>James Franco</name>
            <ssn>123-45-6789</ssn>        
        </phi>    
        <result>The test passed</result>
    </test>
    <test name="my test 2">
        <phi>
            <name>Mister T</name>
            <ssn>123-55-5555</ssn>        
        </phi>    
        <result>The test failed!</result>
    </test>
    <test name="my test 3">
        <phi>
            <name>Ronald Rump</name>
            <ssn>666-66-6666</ssn>        
        </phi>    
        <result>The test result is unknown</result>
    </test>
</root>
```

We can replace all the phi nodes with nothing :
```bash
$ stringaling replaceall -i results.xml -o clean-results.xml -s '<phi>' -e '</phi>'
```

Giving us the resulting `clean-results.xml`:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<root>
    <test name="my test">
      
        <result>The test passed</result>
    </test>
    <test name="my test 2">
      
        <result>The test failed!</result>
    </test>
    <test name="my test 3">
      
        <result>The test result is unknown</result>
    </test>
</root>
```

#### Combine
This command combines a set of text files into a single file.

```bash
$ stringaling combine|c [-v] -f FILE [-f FILE]... -o OUTPUT_FILE [-d]
``` 

The command can either be `combine` or `c` for short.

##### Minimum Requirements
You need at least 1mb of free memory to run this command.

##### Arguments
* -f FILE
  * A file to combine, this option can be supplied multiple times
* -o OUTPUT_FILE
  * The file to write the combination to
* -d
  * When supplied, will delete the input files after combination