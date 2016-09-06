## Postfix log spam analyzer

This service is useful for none paranoid mail system which adds SPAM headers to the mail in contrast to the rejects or quarantine. Works as stand along process and tails mail.log file. The goal is to gather information: client address, mail from, spamassassin marker, count spam victims per mail thread. Default and simple reaction is write entry to the own log file on dirty mail, advanced - insert sql record. The second way is prefered in complex with postfix client_acces and sender_access rules

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
        AND `created` >= NOW() - INTERVAL 20 DAY
        GROUP BY `client`
            HAVING (1 - POW(EXP(1), -(`vsum`/20))) > 0.1
    ) `sp`
```

Rejection probability can be caculated as spam rate with daily incidence `1 - POW(EXP(1), -(vsum / 20))`. The value will grow on each spam attemp according to the 20 days period.
