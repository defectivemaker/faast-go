To run, provide your yaml file as the first argument
git clone ...
cd faast-go
go build cmd/main
./main my-yaml-config.yml

Sample yaml config

```
    # currently there is only a payload option. In the future there will be subdomain and file enumeration
    type: payload
    endpoint: https://example.com
    # validateType can be size or code.
    # validateType: size means that successful results are the responses that are not 0 bytes
    # validateType: code means the successful results are the responses that are not 404's
    validateType: size # this means that it will only print out results that are not size 0
    # send cookies with your request
    cookies:
        - sdf=asd
        - asdf=dsfa
    # the number of fields must be the same as the number of wordlists + staticValues
    # field 1 will match with wordlist 1, field 2 with worldlist 2, after all the wordlists are
    # linked with a field, the staticValues will link with a field
    fields:
        - username
        - password
        - extra_field
    wordlists:
        - lists/names-list.txt
        - lists/xato-net-10-million-passwords.txt
    staticValues:
        - static_val
```

The above config will send payload

`https://example.com?username=brian&password=123456&extra_field=static_val`

where brian is in the first line of `lists/names-list.txt` and 123456 is the first
line in `lists/xato-net-10-million-passwords.txt`
