# StructureChest
A structure manager for Minecraft Bedrock Edition.

A work in progress; this is not even close to being done, don't try to use it.

## RoadMap:

To add to app as a whole:
- Settings page (change default directories, change ignored structure prefixes)

To add to the structures tab:
- Be able to delete, move, and rename structures from inside the tree
- Use drag'n'drop for importing structures into worlds
- Tree selection that works like in a GUI file explorer (ctrl for multi, shift for range, etc)
- Double-clicking a world opens a detailed view of the structures inside itl should allow you to drag a structure to the tree to save it to the world or copy it into a pack in the world.
- Double-clicking a structure file opens a detailed view window for it, with a rendered preview (maybe some stuff like layer view, inventory preview, image export?)
- Show world thumbnail, if one exists, instead of the folder icon.
- Export selected structures as zip, regular downloads, or a schematic pack (currently there is permission for HoloPrint and Construct)
- allow opening worlds to the detailed from a selected folder instead of from the com.mojang folder, maybe even for .mcworld files since theyre just a zip of the normal format

To add to the worlds tab:
- List worlds in com.mojang folder, similarly to what already exists in structures tab, just without the detailed structure and pack info.
- Option to rename worlds, open folder location, delete world
- Export world as .mcworld file
- Open world to map view, show world from above; should allow world trimming and dimension deletion (take some inspiration from Chunker's ui here)
- Include some debug tools like listing chunks with enormous amounts of entites, a high chunk file size, etc. 
- Might be possible to include a pending tick cleaerer here, if we do the NBT stuff required to do so.
- two-pane view with option to move selected chunks on the left into a specified location in the world on the right; Should show a preview of the transfer

To add to the packs tab:
- Pack repository similar to the structure tree, but should have a visual folder that's just a mirror of the client's imported/cached packs
- Show information about each pack, like the version(s) it supports, description of the pack itself, etc
- World list similar to the structure tab but shows better information about the packs in each world
- Add/remove packs from worlds
- Possible system to update all instances of a pack to a new one, can implement without internet access if its just the user telling it what pack to replace with what, but adding an optional field for a Github or CurseForge link to download updated versions from (with similar functionality to the Prism Launcher for JE) is a thought.

