---
title: League Export & Import
---

# League Export & Import

League Transfer lets you package up an entire SMB4 league, every team and every player, and hand it to someone else so they can load it into their own game. This is what makes community leagues possible: one person (or a group) puts together a custom league, exports it, and shares it so others can play with the same teams.

League Transfer is a separate mode from the franchise tracker — choose it from the mode picker when you launch smb-tools. It works independently of any franchise you've set up for tracking; it operates directly on the league files inside your SMB4 save data.

After import, the recipient's game treats it as a normal league save. They can simulate seasons, make trades, and run the league however they like; smb-tools is only involved in the transfer itself.

## Exporting a League

The Export tab scans your machine for every SMB4 league it can find and lists each one with its conference, division, and team counts. Pick a league and export it — smb-tools packages the league's save data into a single `.zip` file you can send to someone else.

## Importing a League

The Import tab takes an exported `.zip` file and lets you choose which of your SMB4 save profiles to import it into. Before anything is written, smb-tools shows you a preview of what's in the package and flags any target where that league is already registered.

**No existing league save is ever modified or overwritten.** Importing only ever adds new league files alongside what's already there.

Importing a league does require registering it in `master.sav`, the single registry file SMB4 uses to know which leagues exist. smb-tools always backs up `master.sav` — with a timestamped copy that's never overwritten — immediately before making that change, so a previous state is always recoverable.

For safety, smb-tools refuses to import while SMB4 is running.

::: warning Trust your source
smb-tools does not scan imported league files for malware. Only import files from people you trust.
:::

## Compatibility

Both export and import require SMB4.
