# String-a-ling
Stringaling is a streaming string replacement tool for handling string replacement in extremely large files 
in an environment with limited resources.

Its going to take a long time for large files (about 40 minutes with verbosity for a 200mb xml file)
but it will get the job done.

## Expectations
It's important to set expectations for this program/library.
This does NOT support regex at all, and should only be used for known token replacements.

This can run on a raspberry pi zero, and take all day, but it will run and it will complete. 

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
$ stringaling replaceall|rall [-v] -i INPUT_FILE -o OUTPUT_FILE -s START_TOKEN -e END_TOKEN [-w TOKEN] 
``` 

The command can either be `replaceall` or `rall` for short.

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
$ stringaling replaceall -i results.xml -o clean-results.xml -s '<phi>' -e '<//phi>'
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

