---
title: Kafka Notes
date: '2018-12-25'
---

Topics and Partitions

Topic formed of multiple partitions

Topics a particular stream of data
* Similar to a table in a db
* YOu can have as many topics as you want
* Topic is identified by its name

Topics are split in partitions
 * Each partition is ordered
 * Each message within a partition gets an incremental id, called __offset__
 * Each partition have their own offset


* Offset  only have meanining in a partition.
* Order is only guaranteed only within a partition, offsets are independent.
* Default is kept for one week, offsets dissapear.
* Once the data is writter to a partition, it can't be changed.

Brokers

* Kafka cluster is composed of multiple brokers (servers)
* Each broker is identified with its ID (integer)
* Each broker contains certain topic partitions
* After connecting to any broker (called a bootstrap brooker), you will be connected to the entire cluster.


```
[ Broker 1      ] 
[ [Topic 1]     ]
[ [Partition 0] ]
[               ]
[ [Topic 2]     ]
[ [Partition 1] ]

[ Broker 2      ] 
[ [Topic 1]     ]
[ [Partition 2] ]
[               ]
[ [Topic 2]     ]
[ [Partition 0] ]

[ Broker 3      ] 
[ [Topic 1]     ]
[ [Partition 1] ]
[               ]
[               ]
[               ]
```

Not every broker has every topic.