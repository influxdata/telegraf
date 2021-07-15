# Chess Input Plugin

The `chess` plugin gathers metrics from chess.com on specific players, 
such as: online status, general stats, match data, and regional 
distribution of players and clubs.


### Configuration

```toml
[[inputs.chess]]
# A list of profiles for monotoring 
  profiles = ["username1", "username2"]
#track leaderboard
  leaderboard = false
```

### Troubleshooting

Check that the username is spelt correctly. 

### Example Output

```
measurement1,tag1=foo,tag2=bar field1=1i,field2=2.1 1453831884664956455
measurement2,tag1=foo,tag2=bar,tag3=baz field3=1i 1453831884664956455
```
