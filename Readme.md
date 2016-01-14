## Postfix log spam analyzer

This service is usful for none paranoid mail system which adds SPAM headers to the mail in contrast to the rejects or quarantine. Works as stand along process and tails mail.log file. The goal is to gather information: client address, mail from, spamassassin marker, count spam victims per mail thread. Default and simple reaction is write entry to the own log file on dirty mail, advanced - insert sql record. The second way is prefered in complex with postfix client_acces and sender_access rules

### Desined for

- Postfix
- Amavis (Spammassassin)
- Spamd (Spamassassin)

### How to use with postfix

Create MySQL table

```
CREATE TABLE `spammers` (
   `client` varchar(255) NOT NULL COMMENT 'Client address',
   `created` datetime NOT NULL COMMENT 'Record date in the log',
   `spam_victims_score` int(10) NOT NULL DEFAULT '0' COMMENT 'Recipeints count',
   PRIMARY KEY (`client`,`created`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8 COMMENT='Spammers log'
```

- Uncomment in the postlog-sa.ini db section and fix parameters according to the database access settings
- Uncomment sql section and write sql query.

```
INSERT INTO `spammers`(`client`, `created`, `spam_victims_score`) \
    VALUES(?f, ?t, ?s), (?c, ?t, ?s) \
    ON DUPLICATE KEY UPDATE `client` = `client`
```

#### Query arguments

```
?i - mail thread id
?m - message id
?f - field From
?c - client IP
?t - client connection time
?s - recipients count
```

#### Postfix settings

Caution: postfix must support MySQL(http://www.postfix.org/MYSQL_README.html)

Edit postfix/main.cf

```
smtpd_client_restrictions = check_client_access mysql:/etc/postfix/mysql/client_access.cf
smtpd_sender_restrictions = reject_unknown_address,
                     check_sender_access mysql:/etc/postfix/mysql/client_access.cf
```

Write /etc/postfix/mysql/client_access.cf

```
user = db_user
password = db_user_pass
dbname = db_name
select_field = access
expansion_limit = 1
hosts = 127.0.0.1
query = SELECT "REJECT" `access`
    FROM (SELECT COUNT(*) `hope`, `client`, SUM(spam_victims_score) `vsum`, MAX(`created`) `c`
        FROM `spammers` WHERE `client` = '%s'
        GROUP BY `client`
            HAVING POW(EXP(1), -(`hope`/`vsum`)) > 0.6
                AND (NOW() - INTERVAL 5 DAY) < `c`
    ) `sp`
```

`POW(EXP(1), -(hope/vsum))` is stability boundary to have of the variable trigger rejection. The result will grow on each spam attemp according to the mail recipients. Be careful to compare result from 0 because there can be fault spam alarm made by spamassassin and You will get immidiate reaction. To forgive old, use interval comparison from the last record date.

