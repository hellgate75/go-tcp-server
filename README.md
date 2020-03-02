<p align="center">
<image width="150" height="50" src="images/kube-go.png"></image>&nbsp;
<image width="260" height="410" src="images/golang-logo.png">
&nbsp;<image width="150" height="50" src="images/tls-logo.png"></image>
</p><br/>
<br/>

# Go TCP Server

Go TLS TCP Client / Server library

## Goals

Definition of a secure layer transfer channel using TLS Serve/Client PEM certificates/keys only protocol. 


## Reference Repositories

* [Go TCP Client](https://github.com/hellgate75/go-tcp-client) Client side application

## How does it work?

Server starts with one or more input server certificate/key pairs. 

Call ```help``` ```--help``` or ```-h``` from command line to print out the Server command help.

Source test script is :
```
[start-server.sh](/start-server.sh)
```
It accepts optional parameters or the help request.


## Available plug-ins

Here a sample screen of commands help request:

Available commands are:

* transfer-file: transfers files, create folders and copy folder files into to the remote server

* shell: can run interactively (command prompt, without any parameter) or can run not interactively commands and script remotely, copying script files remotely, executing them and deleting the remote files at the end of the execution.



## Need sample certificates?

You can produce test client/server certificates using following provided script:

```
makecert.sh [-d] or using any of provided parameters. 
```

Run it without any parameter to read the usage.



## Build server

Build server:

```
go install -v -gcflags "-N -l" github.com/hellgate75/go-tcp-server
```



## Get client and server

Get server:

```
go get -u github.com/hellgate75/go-tcp-server
```

Get client:

```
go get -u github.com/hellgate75/go-tcp-client
```


Enjoy the experience.



## License

The library is licensed with [LGPL v. 3.0](/LICENSE) clauses, with prior authorization of author before any production or commercial use. Use of this library or any extension is prohibited due to high risk of damages due to improper use. No warranty is provided for improper or unauthorized use of this library or any implementation.

Any request can be prompted to the author [Fabrizio Torelli](https://www.linkedin.com/in/fabriziotorelli) at the following email address:

[hellgate75@gmail.com](mailto:hellgate75@gmail.com)




