# Flash
A simple, zero-dependency TCP server written in Go for Monitoring and Ticketing Speed Violations

## Specification:
[Protohackers - Speed Daemon](https://protohackers.com/problem/6)

## What it does?
- 🚗 **Manages TCP connections** for speed cameras & ticket dispatchers.  
- 🔍 **Tracks vehicle movements** via license plate observations.  
- 📏 **Calculates average speed** between camera checkpoints.  
- 🎟️ **Issues speeding tickets** if speed limit is exceeded.  
- 📅 **Enforces one ticket per car per day** rule.  
- 📂 **Stores undelivered tickets** until a dispatcher connects.  
- ⚡ **Supports 150+ concurrent clients** efficiently.  

## Why It's Cool
- **Zero dependency**: Pure Go `stdlib` 
- **Fearless Concurrency**: Handles over 200 concurrent cameras and dispatchers with ease.
- **Extremely Lightweight**: Custom binary protocol for minimal overhead.
- **Tiny**: One Executable

## Requirements
- Go 1.21+

