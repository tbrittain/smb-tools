---
title: League Export & Import
---

# League Export & Import

League Transfer lets you package up an entire SMB4 league, every team and every player, and hand it to someone else so they can load it into their own game. This is what makes community leagues possible: one person (or a group) puts together a custom league, exports it, and shares it so others can play with the same teams.

League Transfer is a separate mode from the franchise tracker. Choose it from the mode picker when you launch smb-tools. It works independently of any franchise you've set up for tracking; it operates directly on the league files inside your SMB4 save data.

After import, the recipient's game treats it as a normal league save. They can simulate seasons, make trades, and run the league however they like; smb-tools is only involved in the transfer itself.

## Exporting a League

The Export tab scans your machine for every SMB4 league it can find and lists each one with its conference, division, and team counts. Upon picking a league and exporting it, smb-tools packages the league's save data into a single `.zip` file you can send to someone else.

If your save files aren't in the usual Steam location, say you keep them on an external drive or a separate install, use the Browse Folder button to point smb-tools at that folder instead. It scans wherever you choose the same way it scans the default location, and whatever it finds gets added to the list.

::: warning Empty league shell and save game are not the same export
SMB4 tracks a league as two separate pieces, and smb-tools lets you export either one on its own. The empty league shell is the teams, conferences, and divisions setup with no games played. It is what shows up under Customizations in SMB4. A save game, meaning a Franchise, Season, or Elimination run, is the actual in-progress game that was built from that shell.

Exporting a save game does not include the shell. If you send someone a Franchise, Season, or Elimination export, they will get that game, but they will not get a matching Customizations entry for it. smb-tools labels this kind of export "Export Save Game Only" wherever it appears alongside a save game for the same league. If you also want the recipient to have the league available under Customizations, export and send the empty shell separately, labeled "Export Empty League."

If you only export the save game and skip the shell, the recipient can still get the shell back without you sending it separately, by using the "Export to League" option from inside SMB4 once they have the save game loaded. That in-game option recreates a league shell from the save game itself.
:::

## Exporting from a Franchise Snapshot

The Export tab has two sub-tabs: From Save File, which is everything described above, and From Franchise Snapshot. If you're tracking a franchise in smb-tools, every season sync quietly captures a snapshot of that save behind the scenes, and the From Franchise Snapshot sub-tab lets you turn any of those into a shareable league export. That's handy if you want to send someone an earlier point in your franchise's history instead of whatever state it's in today.

Snapshots are listed by franchise, with each one labeled by season number and the date it was captured. Pick the one you want and give it a name. Naming is required for this kind of export, since a snapshot doesn't carry a current league name the way a live save does. The export also gets a brand new league identity, so it won't collide with the franchise it came from, or with itself if you export the same snapshot more than once. Since a snapshot is always a franchise save in progress, its export button reads "Export Save Game Only" too, and the same shell caveat above applies: send the empty league shell separately if you want the recipient to have a matching Customizations entry.

Exporting a snapshot only reads from it so the franchise and its other snapshots are left exactly as they were.

::: info Snapshots provide point-in-time restore capability
Every time you sync a season means a new snapshot is created. This means that for every snapshot you have, you can create and share a league with the exact progress of that franchise from when it was synced.
:::

## Importing a League

The Import tab takes an exported `.zip` file and lets you choose which of your SMB4 save profiles to import it into. Before anything is written, smb-tools shows you a preview of what's in the package and flags any target where that league is already registered.

**No existing league save is ever modified or overwritten.** Importing only ever adds new league files alongside what's already there.

Importing a league does require registering it in `master.sav`, the single registry file SMB4 uses to know which leagues exist. smb-tools always backs up `master.sav` (with a timestamped copy that's never overwritten) immediately before making that change, so a previous state is always recoverable.

For safety, smb-tools refuses to import while SMB4 is running.

::: warning Trust your source
smb-tools does not scan imported league files for malware. Only import files from sources you trust.
:::

## Compatibility

Both export and import require SMB4.
