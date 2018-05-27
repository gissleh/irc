# Protocol Samples

This file contains samples from nc-ing IRC servers, useful as a quick protocol reference.

## Register
Just NICK and USER without any CAP negotiation

```irc
NICK test
:archgisle.lan NOTICE * :*** Checking Ident
:archgisle.lan NOTICE * :*** Looking up your hostname...
:archgisle.lan NOTICE * :*** No Ident response
:archgisle.lan NOTICE * :*** Checking your IP against DNS blacklist
:archgisle.lan NOTICE * :*** Couldn't look up your hostname
:archgisle.lan NOTICE * :*** IP not found in DNS blacklist
USER test 8 * :Test test
:archgisle.lan 001 test :Welcome to the TestServer Internet Relay Chat Network test
:archgisle.lan 002 test :Your host is archgisle.lan[archgisle.lan/6667], running version charybdis-4-rc3
:archgisle.lan 003 test :This server was created Fri Nov 25 2016 at 17:28:20 CET
:archgisle.lan 004 test archgisle.lan charybdis-4-rc3 DQRSZagiloswxz CFILNPQbcefgijklmnopqrstvz bkloveqjfI
:archgisle.lan 005 test FNC SAFELIST ELIST=CTU MONITOR=100 WHOX ETRACE KNOCK CHANTYPES=#& EXCEPTS INVEX CHANMODES=eIbq,k,flj,CFLNPQcgimnprstz CHANLIMIT=#&:15 :are supported by this server
:archgisle.lan 005 test PREFIX=(ov)@+ MAXLIST=bqeI:100 MODES=4 NETWORK=TestServer STATUSMSG=@+ CALLERID=g CASEMAPPING=rfc1459 NICKLEN=30 MAXNICKLEN=31 CHANNELLEN=50 TOPICLEN=390 DEAF=D :are supported by this server
:archgisle.lan 005 test TARGMAX=NAMES:1,LIST:1,KICK:1,WHOIS:1,PRIVMSG:4,NOTICE:4,ACCEPT:,MONITOR: EXTBAN=$,&acjmorsuxz| CLIENTVER=3.0 :are supported by this server
:archgisle.lan 251 test :There are 0 users and 2 invisible on 1 servers
:archgisle.lan 254 test 1 :channels formed
:archgisle.lan 255 test :I have 2 clients and 0 servers
:archgisle.lan 265 test 2 2 :Current local users 2, max 2
:archgisle.lan 266 test 2 2 :Current global users 2, max 2
:archgisle.lan 250 test :Highest connection count: 2 (2 clients) (8 connections received)
:archgisle.lan 375 test :- archgisle.lan Message of the Day - 
:archgisle.lan 372 test :- This server is only for testing irce, not chatting. If you happen
:archgisle.lan 372 test :- to connect to it by accident, please disconnect immediately.
:archgisle.lan 372 test :-  
:archgisle.lan 372 test :-  - #Test  :: Test Channel
:archgisle.lan 372 test :-  - #Test2 :: Other Test Channel
:archgisle.lan 372 test :-  
:archgisle.lan 372 test :- Sopp sopp sopp sopp
:archgisle.lan 376 test :End of /MOTD command.
:test MODE test :+i
```

## CAP negotiation

```irc
CAP LS 302
:archgisle.lan NOTICE * :*** Checking Ident
:archgisle.lan NOTICE * :*** Looking up your hostname...
:archgisle.lan NOTICE * :*** No Ident response
:archgisle.lan NOTICE * :*** Checking your IP against DNS blacklist
:archgisle.lan NOTICE * :*** Couldn't look up your hostname
:archgisle.lan NOTICE * :*** IP not found in DNS blacklist
:archgisle.lan CAP * LS :account-notify account-tag away-notify cap-notify chghost echo-message extended-join invite-notify multi-prefix server-time tls userhost-in-names 
CAP REQ :server-time userhost-in-names
:archgisle.lan CAP * ACK :server-time userhost-in-names
CAP REQ multi-prefix away-notify
:archgisle.lan CAP * ACK :multi-prefix
CAP END
NICK test
USER test 8 * :test
:archgisle.lan 001 test :Welcome to the TestServer Internet Relay Chat Network test

```

## Channel joining
```irc
JOIN #Test
:test!~test@127.0.0.1 JOIN #Test
:archgisle.lan 353 test = #Test :test @Gisle
:archgisle.lan 366 test #Test :End of /NAMES list.
:Gisle!~irce@10.32.0.1 PART #Test :undefined
:Gisle!~irce@10.32.0.1 JOIN #Test
```

## Setting modes
```irc
:test MODE test :+i
JOIN #Test
:Testify!~test@127.0.0.1 JOIN #Test
:archgisle.lan 353 Testify = #Test :Testify @Gisle
:archgisle.lan 366 Testify #Test :End of /NAMES list.
:Gisle!~irce@10.32.0.1 MODE #Test +osv Testify Testify
:Gisle!~irce@10.32.0.1 MODE #Test +N-s 
```