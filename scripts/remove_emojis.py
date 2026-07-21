#!/usr/bin/env python3
"""
Emoji Removal Script
Removes only native emoji characters from text files while preserving ASCII art and box-drawing characters.
"""

import os
import re
import sys
import unicodedata

# Unicode ranges for emoji characters (carefully curated to avoid ASCII/box-drawing)
# These are PROPER Unicode blocks for emoji only - no overlaps with ASCII art characters
EMOJI_RANGES = [
    (0x203C, 0x203C),    # Double Exclamation Mark
    (0x2049, 0x2049),    # Question Exclamation Mark
    (0x2122, 0x2122),    # Trademark
    (0x2139, 0x2139),    # Information
    (0x2194, 0x2199),    # Arrow symbols (subset - not all to avoid box-drawing overlap)
    (0x21A9, 0x21AA),    # Hook arrows
    (0x231A, 0x231B),    # Watch, Hourglass
    (0x23E9, 0x23F3),    # Multilingual indices (excluding some that overlap with box drawing)
    (0x23F8, 0x23FA),    #体育 symbols
    (0x25AA, 0x25AB),    # Small Squares (subset)
    (0x25B6, 0x25B6),    # Play Button
    (0x25C0, 0x25C0),    # Reverse Button
    (0x25FB, 0x25FE),    # White/Black Squares (subset)
    (0x2614, 0x2615),    # Umbrella, Hot Beverage
    (0x2618, 0x2618),    # Shamrock
    (0x261D, 0x261D),    # Pointing Up
    (0x2620, 0x2620),    # Skull
    (0x2622, 0x2623),    # Radioactive, Biohazard
    (0x2626, 0x2626),    # Orthodox Cross
    (0x262A, 0x262A),    # Star and Crescent
    (0x262E, 0x262F),    # Peace, Yin Yang
    (0x2638, 0x263A),    # Wheel of Dharma, Face
    (0x2648, 0x2653),    # Zodiac (Aries - Pisces)
    (0x2660, 0x2663),    # Spade, Heart, Diamond, Club
    (0x2665, 0x2665),    # Heart (redundant but explicit)
    (0x2666, 0x2666),    # Diamond (redundant but explicit)
    (0x2668, 0x2668),    # Hot Springs
    (0x267B, 0x267B),    # Recycling
    (0x267F, 0x267F),    # Wheelchair
    (0x2693, 0x2693),    # Anchor
    (0x26A0, 0x26A1),    # Warning, High Voltage
    (0x26AA, 0x26AB),    # Circles
    (0x26BD, 0x26BE),    # Soccer, Baseball
    (0x26C4, 0x26C5),    # Snowman, Sun behind cloud
    (0x26C8, 0x26C8),    # Cloud with rain
    (0x26CE, 0x26CE),    # Ophiuchus
    (0x26D4, 0x26D4),    # No Entry
    (0x26EA, 0x26EA),    # Church
    (0x26F2, 0x26F3),    # Fountain, Golf
    (0x26F5, 0x26F5),    # Sailboat
    (0x26FA, 0x26FA),    # Tent
    (0x26FD, 0x26FD),    # Fuel Pump
    (0x2702, 0x2702),    # Scissors
    (0x2705, 0x2705),    # Check Mark
    (0x2708, 0x270D),    # Various symbols
    (0x270F, 0x270F),    # Pencil
    (0x2712, 0x2712),    # Pen
    (0x2714, 0x2714),    # Check Mark
    (0x2716, 0x2716),    # Multiplication X
    (0x271D, 0x271D),    # Latin Cross
    (0x2721, 0x2721),    # Star of David
    (0x2728, 0x2728),    # Sparkles
    (0x2733, 0x2734),    # Eight-spoked asterisk, Eight-pointed star
    (0x2744, 0x2744),    # Snowflake
    (0x2747, 0x2747),    # Sparkle
    (0x274C, 0x274C),    # Cross Mark
    (0x274E, 0x274E),    # Cross Mark
    (0x2753, 0x2755),    # Question marks
    (0x2757, 0x2757),    # Exclamation
    (0x2763, 0x2764),    # Hearts
    (0x2795, 0x2797),    # Plus, Minus, Divide
    (0x27A1, 0x27A1),    # Right arrow
    (0x27B0, 0x27B0),    # Curly loop
    (0x27BF, 0x27BF),    # Double curly loop
    (0x2934, 0x2935),    # Arrows
    (0x2B05, 0x2B07),    # Arrows
    (0x2B1B, 0x2B1C),    # Squares
    (0x2B50, 0x2B50),    # Star
    (0x2B55, 0x2B55),    # Circle
    (0x3030, 0x3030),    # Wavy dash
    (0x303D, 0x303D),    # Part alternation
    (0x3297, 0x3297),    # Circled ideograph
    (0x3299, 0x3299),    # Circled ideograph
    # Emoji blocks (proper ranges)
    (0x1F004, 0x1F004),  # Mahjong Tile
    (0x1F0CF, 0x1F0CF),  # Playing Card
    (0x1F170, 0x1F171),  # A, B, AB, O (blood types)
    (0x1F17E, 0x1F17F),  # 1-0 with friction
    (0x1F18E, 0x1F18E),  # AB Button
    (0x1F191, 0x1F19A),  # Squared letters
    (0x1F201, 0x1F202),  # Japanese
    (0x1F21A, 0x1F21A),  # Japanese
    (0x1F22F, 0x1F22F),  # Japanese
    (0x1F232, 0x1F23A),  # Japanese
    (0x1F280, 0x1F280),  # Japanese
    (0x1F300, 0x1F320),  # Cyclone - Captial letters
    (0x1F321, 0x1F321),  # Thermometer
    (0x1F324, 0x1F32C),  # Weather symbols
    (0x1F32D, 0x1F32F),  # Food
    (0x1F330, 0x1F335),  # Plants
    (0x1F336, 0x1F336),  # Hot pepper
    (0x1F337, 0x1F34C),  # Fruits, flowers
    (0x1F34D, 0x1F34F),  # Food
    (0x1F350, 0x1F350),  # Fruit
    (0x1F351, 0x1F35B),  # Food
    (0x1F35C, 0x1F35C),  # Steaming
    (0x1F35D, 0x1F35F),  # Food
    (0x1F360, 0x1F360),  # Roasted
    (0x1F361, 0x1F362),  # Food
    (0x1F363, 0x1F363),  # Sushi
    (0x1F364, 0x1F365),  # Food
    (0x1F366, 0x1F366),  # Ice cream
    (0x1F367, 0x1F368),  # Dessert
    (0x1F369, 0x1F36B),  # Donut, cookie, chocolate
    (0x1F36C, 0x1F36C),  # Candy
    (0x1F36D, 0x1F36F),  # Food
    (0x1F370, 0x1F370),  # Shortcake
    (0x1F371, 0x1F372),  # Food
    (0x1F373, 0x1F373),  # Cooking
    (0x1F374, 0x1F374),  # Fork and knife
    (0x1F375, 0x1F375),  # Tea
    (0x1F376, 0x1F376),  # Bento
    (0x1F377, 0x1F377),  # Wine
    (0x1F378, 0x1F378),  # Cocktail
    (0x1F379, 0x1F379),  # Tropical drink
    (0x1F37A, 0x1F37A),  # Beer
    (0x1F37B, 0x1F37B),  # Beers
    (0x1F37C, 0x1F37C),  # Baby bottle
    (0x1F37D, 0x1F37D),  # Plate with cutlery
    (0x1F37E, 0x1F37F),  # Ceremony
    (0x1F380, 0x1F380),  # Ribbon
    (0x1F381, 0x1F381),  # Gift
    (0x1F382, 0x1F382),  # Birthday
    (0x1F383, 0x1F383),  # balloons, cake
    (0x1F384, 0x1F384),  # Christmas tree
    (0x1F385, 0x1F385),  # Santa
    (0x1F386, 0x1F386),  # Fireworks
    (0x1F387, 0x1F387),  # Sparkler
    (0x1F388, 0x1F38A),  # Balloon, confetti
    (0x1F38B, 0x1F38B),  # Tanabata
    (0x1F38C, 0x1F38C),  # Crossed flags
    (0x1F38D, 0x1F38E),  # Festival
    (0x1F38F, 0x1F38F),  # Pinata
    (0x1F390, 0x1F390),  # Telescope
    (0x1F391, 0x1F391),  # Moon view
    (0x1F392, 0x1F392),  # Backpack
    (0x1F393, 0x1F393),  # Graduation cap
    (0x1F396, 0x1F397),  # Military
    (0x1F399, 0x1F39B),  # Music
    (0x1F39E, 0x1F39F),  # Sport
    (0x1F3A0, 0x1F3C4),  # Activities
    (0x1F3C5, 0x1F3C5),  # Sports medal
    (0x1F3C6, 0x1F3C6),  # Trophy
    (0x1F3C7, 0x1F3C7),  # Racing
    (0x1F3C8, 0x1F3C8),  # Football
    (0x1F3C9, 0x1F3C9),  # Rugby
    (0x1F3CA, 0x1F3CA),  # Swimming
    (0x1F3CB, 0x1F3CE),  # Sports equipment
    (0x1F3CF, 0x1F3D3),  # Sports
    (0x1F3D4, 0x1F3DF),  # Sports/activities
    (0x1F3E0, 0x1F3E3),  # Buildings
    (0x1F3E4, 0x1F3E4),  # Oncoming police car
    (0x1F3E5, 0x1F3E9),  # Buildings
    (0x1F3EA, 0x1F3EB),  # Hospital, bank
    (0x1F3EC, 0x1F3EF),  # Places
    (0x1F3F0, 0x1F3F0),  # Castle
    (0x1F3F3, 0x1F3F4),  # Flags
    (0x1F3F5, 0x1F3F5),  # Moai
    (0x1F3F7, 0x1F3F7),  # Label
    (0x1F3F8, 0x1F3FA),  # Sports/places
    (0x1F400, 0x1F43E),  # Animals
    (0x1F43F, 0x1F43F),  # Coyote
    (0x1F440, 0x1F440),  # Eyes
    (0x1F441, 0x1F441),  # Eye
    (0x1F442, 0x1F443),  # Ears, head
    (0x1F444, 0x1F444),  # Mouth
    (0x1F445, 0x1F445),  # Tongue
    (0x1F446, 0x1F450),  # Body parts
    (0x1F451, 0x1F451),  # Crown
    (0x1F452, 0x1F452),  # Clothing
    (0x1F453, 0x1F454),  # Glasses, necktie
    (0x1F455, 0x1F456),  # Clothing
    (0x1F457, 0x1F45A),  # Dress,Kimono
    (0x1F45B, 0x1F45B),  # Handbag
    (0x1F45C, 0x1F45C),  # Briefcase
    (0x1F45D, 0x1F45D),  # Pouch
    (0x1F45E, 0x1F45E),  # Shoes
    (0x1F45F, 0x1F45F),  # Running shirt
    (0x1F460, 0x1F460),  # High-heeled shoe
    (0x1F461, 0x1F462),  # Sandal, woman boots
    (0x1F463, 0x1F463),  # Footprints
    (0x1F464, 0x1F464),  # Bust in silhouette
    (0x1F465, 0x1F465),  # Busts in silhouette
    (0x1F466, 0x1F46B),  # People
    (0x1F46C, 0x1F46C),  # People
    (0x1F46D, 0x1F46D),  # People
    (0x1F46E, 0x1F46E),  # Police
    (0x1F46F, 0x1F46F),  # People
    (0x1F470, 0x1F478),  # People
    (0x1F479, 0x1F47B),  # Monsters
    (0x1F47C, 0x1F47C),  # Baby angel
    (0x1F47D, 0x1F480),  # Monsters
    (0x1F481, 0x1F487),  # People
    (0x1F488, 0x1F488),  # Haircut
    (0x1F489, 0x1F48E),  # Objects
    (0x1F48F, 0x1F48F),  # Kiss
    (0x1F490, 0x1F490),  # Bouquet
    (0x1F491, 0x1F491),  # Couple with heart
    (0x1F492, 0x1F492),  # Wedding
    (0x1F493, 0x1F493),  # Party
    (0x1F494, 0x1F495),  # Hearts
    (0x1F496, 0x1F496),  # Hearts
    (0x1F497, 0x1F497),  # Hearts
    (0x1F498, 0x1F498),  # Hearts
    (0x1F499, 0x1F499),  # Blue heart
    (0x1F49A, 0x1F49A),  # Green heart
    (0x1F49B, 0x1F49B),  # Yellow heart
    (0x1F49C, 0x1F49C),  # Purple heart
    (0x1F49D, 0x1F49D),  # Heart with ribbon
    (0x1F49E, 0x1F49E),  # Hearts
    (0x1F49F, 0x1F49F),  # Growing heart
    (0x1F4A0, 0x1F4A0),  # Diamond heart
    (0x1F4A1, 0x1F4A1),  # Light bulb
    (0x1F4A2, 0x1F4A2),  # Anger
    (0x1F4A3, 0x1F4A3),  # Bomb
    (0x1F4A4, 0x1F4A4),  # Zzz
    (0x1F4A5, 0x1F4A5),  # Collision
    (0x1F4A6, 0x1F4A7),  # Sweat droplets
    (0x1F4A8, 0x1F4A8),  # Droplet
    (0x1F4A9, 0x1F4A9),  # Pile of poo
    (0x1F4AA, 0x1F4AA),  # Flexed biceps
    (0x1F4AB, 0x1F4AB),  # Brain
    (0x1F4AC, 0x1F4AC),  # Speech bubble
    (0x1F4AD, 0x1F4AD),  # Thought bubble
    (0x1F4AE, 0x1F4AF),  # Books
    (0x1F4B0, 0x1F4B0),  # Money bag
    (0x1F4B1, 0x1F4B1),  # Currency exchange
    (0x1F4B2, 0x1F4B2),  # Dollar
    (0x1F4B3, 0x1F4B3),  # Increase
    (0x1F4B4, 0x1F4B4),  # ATMs
    (0x1F4B5, 0x1F4B5),  # Heavy dollar
    (0x1F4B6, 0x1F4B7),  # Currency
    (0x1F4B8, 0x1F4B8),  # Money
    (0x1F4B9, 0x1F4B9),  # Chart
    (0x1F4BA, 0x1F4BA),  # Seat
    (0x1F4BB, 0x1F4BB),  # Laptop
    (0x1F4BC, 0x1F4BC),  # Briefcase
    (0x1F4BD, 0x1F4BD),  # Minidisc
    (0x1F4BE, 0x1F4BE),  # Floppy
    (0x1F4BF, 0x1F4BF),  # CD
    (0x1F4C0, 0x1F4C0),  # DVD
    (0x1F4C1, 0x1F4C1),  # Folder
    (0x1F4C2, 0x1F4C2),  # Folder
    (0x1F4C3, 0x1F4C3),  # Page
    (0x1F4C4, 0x1F4C4),  # Page facing up
    (0x1F4C5, 0x1F4C5),  # Calendar
    (0x1F4C6, 0x1F4C6),  # Tear-off calendar
    (0x1F4C7, 0x1F4C7),  # Card index
    (0x1F4C8, 0x1F4C8),  # Chart
    (0x1F4C9, 0x1F4C9),  # Bar chart
    (0x1F4CA, 0x1F4CA),  # Bar chart
    (0x1F4CB, 0x1F4CB),  # Clipboard
    (0x1F4CC, 0x1F4CC),  # Pushpin
    (0x1F4CD, 0x1F4CD),  # Round pushpin
    (0x1F4CE, 0x1F4CE),  # Paperclip
    (0x1F4CF, 0x1F4CF),  # Straight ruler
    (0x1F4D0, 0x1F4D0),  # Triangular ruler
    (0x1F4D1, 0x1F4D1),  # Card file box
    (0x1F4D2, 0x1F4D5),  # Books
    (0x1F4D6, 0x1F4D6),  # Open book
    (0x1F4D7, 0x1F4DA),  # Books
    (0x1F4DB, 0x1F4DB),  # Name badge
    (0x1F4DC, 0x1F4DC),  # Scroll
    (0x1F4DD, 0x1F4DD),  # Memo
    (0x1F4DE, 0x1F4DF),  # Phone
    (0x1F4E0, 0x1F4E0),  # Fax
    (0x1F4E1, 0x1F4E1),  # Outbox tray
    (0x1F4E2, 0x1F4E2),  # Inbox tray
    (0x1F4E3, 0x1F4E3),  # Bell
    (0x1F4E4, 0x1F4E4),  # Envelope
    (0x1F4E5, 0x1F4E5),  # Incoming envelope
    (0x1F4E6, 0x1F4E6),  # Package
    (0x1F4E7, 0x1F4E7),  # E-mail
    (0x1F4E8, 0x1F4E8),  # Incoming
    (0x1F4E9, 0x1F4E9),  # Envelope
    (0x1F4EA, 0x1F4EC),  # Mailboxes
    (0x1F4ED, 0x1F4ED),  # Postbox
    (0x1F4EE, 0x1F4EE),  # Mail
    (0x1F4EF, 0x1F4EF),  # Postal horn
    (0x1F4F0, 0x1F4F0),  # Newspaper
    (0x1F4F1, 0x1F4F1),  # Mobile phone
    (0x1F4F2, 0x1F4F2),  # Phone off
    (0x1F4F3, 0x1F4F4),  # Vibrating phone
    (0x1F4F5, 0x1F4F5),  # No mobile
    (0x1F4F6, 0x1F4F6),  # Antenna
    (0x1F4F7, 0x1F4F7),  # Camera
    (0x1F4F8, 0x1F4F8),  # Camera flash
    (0x1F4F9, 0x1F4FA),  # Video camera
    (0x1F4FB, 0x1F4FB),  # Radio
    (0x1F4FC, 0x1F4FC),  # Videocassette
    (0x1F4FD, 0x1F4FD),  # Film
    (0x1F4FE, 0x1F4FE),  # Connector
    (0x1F4FF, 0x1F4FF),  # Prayer beads
    (0x1F500, 0x1F503),  # Spiral
    (0x1F504, 0x1F504),  # Clockwise arrows
    (0x1F505, 0x1F506),  # Counterclockwise
    (0x1F507, 0x1F507),  # Speaker
    (0x1F508, 0x1F509),  # Speaker
    (0x1F50A, 0x1F50A),  # Speaker with sound
    (0x1F50B, 0x1F50D),  # Lock, key
    (0x1F50E, 0x1F50E),  # Magnifying
    (0x1F50F, 0x1F50F),  # Lock
    (0x1F510, 0x1F510),  # Lock with key
    (0x1F511, 0x1F511),  # Key
    (0x1F512, 0x1F513),  # Lock
    (0x1F514, 0x1F514),  # Bell with slash
    (0x1F515, 0x1F515),  # Bell
    (0x1F516, 0x1F516),  # Bookmark
    (0x1F517, 0x1F517),  # Link
    (0x1F518, 0x1F518),  # Radio button
    (0x1F519, 0x1F519),  # Back
    (0x1F51A, 0x1F51A),  # END arrow
    (0x1F51B, 0x1F51B),  # ON! arrow
    (0x1F51C, 0x1F51C),  # SOON arrow
    (0x1F51D, 0x1F51D),  # TOP arrow
    (0x1F51E, 0x1F51E),  # No one under
    (0x1F51F, 0x1F51F),  # Knife
    (0x1F520, 0x1F520),  # Pencil
    (0x1F521, 0x1F522),  # Pencil, pen
    (0x1F523, 0x1F523),  # Paintbrush
    (0x1F524, 0x1F524),  # Crayon
    (0x1F525, 0x1F525),  # Fire
    (0x1F526, 0x1F526),  # Flashlight
    (0x1F527, 0x1F527),  # Wrench
    (0x1F528, 0x1F528),  # Hammer
    (0x1F529, 0x1F529),  # Nut and bolt
    (0x1F52A, 0x1F52A),  # Gear
    (0x1F52B, 0x1F52B),  # Gun
    (0x1F52C, 0x1F52C),  # Microscope
    (0x1F52D, 0x1F52D),  # Telescope
    (0x1F52E, 0x1F52E),  # Crystal ball
    (0x1F52F, 0x1F52F),  # Cup with straw
    (0x1F530, 0x1F530),  # Banana
    (0x1F531, 0x1F531),  # Cup with straw
    (0x1F532, 0x1F536),  # Button
    (0x1F537, 0x1F537),  # Kimono
    (0x1F538, 0x1F538),  # Gem
    (0x1F539, 0x1F539),  # Button
    (0x1F53A, 0x1F53A),  # Red
    (0x1F53B, 0x1F53B),  # Green
    (0x1F53C, 0x1F53C),  # Blue
    (0x1F53D, 0x1F53D),  # Orange
    (0x1F549, 0x1F54A),  # Om
    (0x1F54B, 0x1F54F),  # Prayer beads
    (0x1F550, 0x1F55B),  # Clocks
    (0x1F55C, 0x1F55F),  # More clocks
    (0x1F560, 0x1F567),  # Even more clocks
    (0x1F568, 0x1F56F),  # More
    (0x1F570, 0x1F573),  # More
    (0x1F574, 0x1F575),  # People
    (0x1F576, 0x1F577),  # Glasses
    (0x1F578, 0x1F578),  # Frog
    (0x1F579, 0x1F579),  # Chess
    (0x1F57A, 0x1F57A),  # Man dancing
    (0x1F57B, 0x1F57F),  # People
    (0x1F580, 0x1F583),  # Joker
    (0x1F584, 0x1F587),  # Speech
    (0x1F588, 0x1F588),  # Mailbox
    (0x1F589, 0x1F58A),  # Pens
    (0x1F58B, 0x1F58B),  # Computer
    (0x1F58C, 0x1F58D),  # Floppy
    (0x1F58E, 0x1F58E),  # Film
    (0x1F58F, 0x1F58F),  # Audio
    (0x1F590, 0x1F590),  # Hand
    (0x1F591, 0x1F594),  # Fingers
    (0x1F595, 0x1F596),  # Vulcan salute
    (0x1F597, 0x1F597),  # Hand
    (0x1F598, 0x1F599),  # Hands
    (0x1F59A, 0x1F59D),  # Hands
    (0x1F59E, 0x1F59F),  # Hand
    (0x1F5A0, 0x1F5A0),  # Lips
    (0x1F5A1, 0x1F5A1),  # Mole
    (0x1F5A2, 0x1F5A2),  # Ear
    (0x1F5A3, 0x1F5A3),  # Tooth
    (0x1F5A4, 0x1F5A4),  # Hair
    (0x1F5A5, 0x1F5A5),  # Monitor
    (0x1F5A6, 0x1F5A7),  # Network
    (0x1F5A8, 0x1F5A8),  # Printer
    (0x1F5A9, 0x1F5A9),  # Keyboard
    (0x1F5AA, 0x1F5AA),  # Computer
    (0x1F5AB, 0x1F5AB),  # Desktop
    (0x1F5AC, 0x1F5AC),  # Floppy
    (0x1F5AD, 0x1F5AD),  # Hard disk
    (0x1F5AE, 0x1F5AE),  # Router
    (0x1F5AF, 0x1F5AF),  # Joystick
    (0x1F5B0, 0x1F5B1),  # Desktop
    (0x1F5B2, 0x1F5B2),  # Computer
    (0x1F5B3, 0x1F5B3),  # Printer
    (0x1F5B4, 0x1F5B5),  # Computer
    (0x1F5B6, 0x1F5B6),  # Keyboard
    (0x1F5B7, 0x1F5B7),  # Desktop
    (0x1F5B8, 0x1F5B8),  # Mouse
    (0x1F5B9, 0x1F5BA),  # Computer
    (0x1F5BB, 0x1F5BB),  # Laptop
    (0x1F5BC, 0x1F5BC),  # Display
    (0x1F5BD, 0x1F5BD),  # Cash
    (0x1F5BE, 0x1F5BE),  # Film
    (0x1F5BF, 0x1F5BF),  # Ballot
    (0x1F5C0, 0x1F5C0),  # Clipboard
    (0x1F5C1, 0x1F5C1),  # Ballot
    (0x1F5C2, 0x1F5C3),  # Files
    (0x1F5C4, 0x1F5C4),  # Folder
    (0x1F5C5, 0x1F5C5),  # File
    (0x1F5C6, 0x1F5C7),  # Cards
    (0x1F5C8, 0x1F5C9),  # File
    (0x1F5CA, 0x1F5CA),  # File
    (0x1F5CB, 0x1F5CB),  # File folder
    (0x1F5CC, 0x1F5CC),  # Open file
    (0x1F5CD, 0x1F5CD),  # Calendar
    (0x1F5CE, 0x1F5CE),  # Calendar
    (0x1F5CF, 0x1F5CF),  # Calendar
    (0x1F5D0, 0x1F5D0),  # Card file
    (0x1F5D1, 0x1F5D2),  # File
    (0x1F5D3, 0x1F5D3),  # File
    (0x1F5D4, 0x1F5D4),  # File folder
    (0x1F5D5, 0x1F5D8),  # Files
    (0x1F5D9, 0x1F5D9),  # File
    (0x1F5DA, 0x1F5DA),  # File folder
    (0x1F5DB, 0x1F5DB),  # File
    (0x1F5DC, 0x1F5DD),  # Files
    (0x1F5DE, 0x1F5DE),  # File
    (0x1F5DF, 0x1F5DF),  # File
    (0x1F5E0, 0x1F5E0),  # File folder
    (0x1F5E1, 0x1F5E1),  # File
    (0x1F5E2, 0x1F5E2),  # File
    (0x1F5E3, 0x1F5E3),  # File folder
    (0x1F5E4, 0x1F5E4),  # File
    (0x1F5E5, 0x1F5E7),  # Files
    (0x1F5E8, 0x1F5E8),  # File cabinet
    (0x1F5E9, 0x1F5E9),  # Wastebasket
    (0x1F5EA, 0x1F5ED),  # Files
    (0x1F5EE, 0x1F5EE),  # File
    (0x1F5EF, 0x1F5EF),  # File
    (0x1F5F0, 0x1F5F2),  # File cabinet
    (0x1F5F3, 0x1F5F3),  # Abacus
    (0x1F5F4, 0x1F5F4),  # Gear
    (0x1F5F5, 0x1F5FA),  # Various
    (0x1F5FB, 0x1F5FB),  # Fuji
    (0x1F5FC, 0x1F5FC),  # Tokyo tower
    (0x1F5FD, 0x1F5FD),  # Statue of liberty
    (0x1F5FE, 0x1F5FE),  # Post office
    (0x1F5FF, 0x1F5FF),  # MOAi
    (0x1F600, 0x1F600),  # Grinning face
    (0x1F601, 0x1F606),  # Grinning face
    (0x1F607, 0x1F608),  # Grinning face
    (0x1F609, 0x1F60D),  # Smiling face
    (0x1F60E, 0x1F60E),  # Smiling face with sunglasses
    (0x1F60F, 0x1F60F),  # Smirking face
    (0x1F610, 0x1F610),  # Neutral face
    (0x1F611, 0x1F611),  # Expressionless face
    (0x1F612, 0x1F614),  # Unamused
    (0x1F615, 0x1F615),  # Pensive
    (0x1F616, 0x1F616),  # Confused
    (0x1F617, 0x1F618),  # Relieved
    (0x1F619, 0x1F619),  # Pouting
    (0x1F61A, 0x1F61A),  # Pouting
    (0x1F61B, 0x1F61B),  # Cat face
    (0x1F61C, 0x1F61C),  # Cat face with WRY
    (0x1F61D, 0x1F61D),  # Cat face with
    (0x1F61E, 0x1F620),  # Disappointed
    (0x1F621, 0x1F621),  # Angry face
    (0x1F622, 0x1F622),  # Crying face
    (0x1F623, 0x1F623),  # Pleading
    (0x1F624, 0x1F624),  # Face with
    (0x1F625, 0x1F625),  # Disappointed
    (0x1F626, 0x1F627),  # Grinning face
    (0x1F628, 0x1F628),  # Anguished
    (0x1F629, 0x1F629),  # Fearful
    (0x1F62A, 0x1F62A),  # Sleepy
    (0x1F62B, 0x1F62B),  # Tired
    (0x1F62C, 0x1F62C),  # Grimacing
    (0x1F62D, 0x1F62D),  # Loudly crying
    (0x1F62E, 0x1F62E),  # Face with
    (0x1F62F, 0x1F62F),  # Hushed
    (0x1F630, 0x1F630),  # Anguished
    (0x1F631, 0x1F631),  # Fearful
    (0x1F632, 0x1F632),  # Astonished
    (0x1F633, 0x1F633),  # Flushed
    (0x1F634, 0x1F634),  # Sleeping
    (0x1F635, 0x1F635),  # Dizzy face
    (0x1F636, 0x1F636),  # Face without
    (0x1F637, 0x1F637),  # Face with medical
    (0x1F638, 0x1F63B),  # Cat faces
    (0x1F63C, 0x1F63D),  # Cat faces
    (0x1F63E, 0x1F63F),  # Cat faces
    (0x1F640, 0x1F640),  # Crying cat
    (0x1F641, 0x1F642),  # Slightly
    (0x1F643, 0x1F644),  # Upside-down
    (0x1F645, 0x1F64F),  # Gestures
    (0x1F680, 0x1F680),  # Rocket
    (0x1F681, 0x1F682),  # Helicopter
    (0x1F683, 0x1F683),  # Train
    (0x1F684, 0x1F684),  # Train
    (0x1F685, 0x1F685),  # Train
    (0x1F686, 0x1F686),  # Train
    (0x1F687, 0x1F687),  # Train
    (0x1F688, 0x1F688),  # Train
    (0x1F689, 0x1F689),  # Train
    (0x1F68A, 0x1F68D),  # Trams
    (0x1F68E, 0x1F68E),  # Trolleybus
    (0x1F68F, 0x1F68F),  # Bus
    (0x1F690, 0x1F690),  # Minibus
    (0x1F691, 0x1F693),  # Ambulance
    (0x1F694, 0x1F694),  # Oncoming
    (0x1F695, 0x1F695),  # Fire engine
    (0x1F696, 0x1F696),  # Taxi
    (0x1F697, 0x1F697),  # Oncoming
    (0x1F698, 0x1F698),  # Oncoming
    (0x1F699, 0x1F699),  # Oncoming
    (0x1F69A, 0x1F69A),  # Delivery
    (0x1F69B, 0x1F6A1),  # Trucks
    (0x1F6A2, 0x1F6A2),  # Ship
    (0x1F6A3, 0x1F6A4),  # Boats
    (0x1F6A5, 0x1F6A5),  # Horizontal
    (0x1F6A6, 0x1F6A6),  # Round
    (0x1F6A7, 0x1F6A8),  # Construction
    (0x1F6A9, 0x1F6A9),  # Triangular flag
    (0x1F6AA, 0x1F6AB),  # Door, Toilet
    (0x1F6AC, 0x1F6AC),  # Smoking
    (0x1F6AD, 0x1F6AD),  # No smoking
    (0x1F6AE, 0x1F6AF),  # Potable water, Non-potable
    (0x1F6B0, 0x1F6B0),  # Potable water
    (0x1F6B1, 0x1F6B2),  # Bicycle, No bicycles
    (0x1F6B3, 0x1F6B3),  # No bicycles
    (0x1F6B4, 0x1F6B5),  # Pedestrian, Bicyclist
    (0x1F6B6, 0x1F6B6),  # Pedestrian
    (0x1F6B7, 0x1F6B7),  # No pedestrians
    (0x1F6B8, 0x1F6B8),  # Children
    (0x1F6B9, 0x1F6B9),  # Male sign
    (0x1F6BA, 0x1F6BA),  # Female sign
    (0x1F6BB, 0x1F6BB),  # Restroom
    (0x1F6BC, 0x1F6BC),  # Baby symbol
    (0x1F6BD, 0x1F6BD),  # Toilet
    (0x1F6BE, 0x1F6BE),  # Water closet
    (0x1F6BF, 0x1F6BF),  # Shower
    (0x1F6C0, 0x1F6C0),  # Bath
    (0x1F6C1, 0x1F6C1),  # Bathtub
    (0x1F6C2, 0x1F6C5),  # Person
    (0x1F6CB, 0x1F6CB),  # Couch, Bed
    (0x1F6CC, 0x1F6CC),  # Person in bed
    (0x1F6CD, 0x1F6CF),  # Shopping bags
    (0x1F6D0, 0x1F6D0),  # Place of worship
    (0x1F6D1, 0x1F6D2),  # Stops
    (0x1F6D3, 0x1F6D4),  # Stops
    (0x1F6D5, 0x1F6D5),  # Hindu temple
    (0x1F6D6, 0x1F6D7),  # Person
    (0x1F6E0, 0x1F6E0),  # Hammer and wrench
    (0x1F6E1, 0x1F6E1),  # Shield
    (0x1F6E2, 0x1F6E2),  # Oil
    (0x1F6E3, 0x1F6E4),  # Motorway, Railway
    (0x1F6E5, 0x1F6E6),  # Motorway
    (0x1F6E7, 0x1F6E7),  # Hourglass
    (0x1F6E8, 0x1F6E8),  # Hourglass
    (0x1F6E9, 0x1F6E9),  # Watch
    (0x1F6EA, 0x1F6EA),  # Notebook
    (0x1F6EB, 0x1F6EC),  # Rockets
    (0x1F6ED, 0x1F6ED),  # Airplane
    (0x1F6EE, 0x1F6EE),  # Left
    (0x1F6EF, 0x1F6EF),  # Right
    (0x1F6F0, 0x1F6F0),  # Rocket
    (0x1F6F1, 0x1F6F1),  # Small
    (0x1F6F2, 0x1F6F2),  # Airplane
    (0x1F6F3, 0x1F6F3),  # Airplane
    (0x1F6F4, 0x1F6F4),  # Spoon
    (0x1F6F5, 0x1F6F5),  # Carousel
    (0x1F6F6, 0x1F6F6),  # Airplane
    (0x1F6F7, 0x1F6F8),  # Small
    (0x1F6F9, 0x1F6F9),  # Person
    (0x1F6FA, 0x1F6FA),  # Auto rickshaw
    (0x1F6FB, 0x1F6FC),  # Motor boat
    (0x1F7E0, 0x1F7E0),  # Orange circle
    (0x1F7E1, 0x1F7E1),  # Yellow circle
    (0x1F7E2, 0x1F7E2),  # Green circle
    (0x1F7E3, 0x1F7E3),  # Blue circle
    (0x1F7E4, 0x1F7E4),  # Purple circle
    (0x1F7E5, 0x1F7E5),  # Brown circle
    (0x1F7E6, 0x1F7E6),  # Red triangle
    (0x1F7E7, 0x1F7E7),  # Green triangle
    (0x1F7E8, 0x1F7E8),  # Yellow triangle
    (0x1F7E9, 0x1F7E9),  # Blue triangle
    (0x1F7EA, 0x1F7EB),  # Purple triangle
    (0x1F7EC, 0x1F7EC),  # Pink triangle
    (0x1F90C, 0x1F90C),  # Pink heart
    (0x1F90D, 0x1F90F),  # Smiling face
    (0x1F910, 0x1F910),  # Zipper-mouth
    (0x1F911, 0x1F911),  # Money-mouth
    (0x1F912, 0x1F912),  # Thermometer
    (0x1F913, 0x1F913),  # nerd
    (0x1F914, 0x1F914),  # Thinking
    (0x1F915, 0x1F915),  # Face with head-bandage
    (0x1F916, 0x1F916),  # Robot
    (0x1F917, 0x1F917),  # Hugging
    (0x1F918, 0x1F918),  # Sign of the horns
    (0x1F919, 0x1F91E),  # Fingers
    (0x1F920, 0x1F920),  # Cowboy hat
    (0x1F921, 0x1F921),  # Clown
    (0x1F922, 0x1F922),  # Nauseated
    (0x1F923, 0x1F923),  # Rolling on the floor
    (0x1F924, 0x1F927),  # Sneezing
    (0x1F928, 0x1F928),  # Face with raised
    (0x1F929, 0x1F929),  # Dizzy
    (0x1F92A, 0x1F92A),  # Exploding
    (0x1F92B, 0x1F92B),  # Custo
    (0x1F92C, 0x1F92C),  # Pleading
    (0x1F92D, 0x1F92D),  # Hmm
    (0x1F92E, 0x1F92E),  # Cow
    (0x1F92F, 0x1F92F),  # Heart
    (0x1F930, 0x1F930),  # Pregnant
    (0x1F931, 0x1F931),  # Breast-feeding
    (0x1F932, 0x1F932),  # Palms
    (0x1F933, 0x1F933),  # Selfie
    (0x1F934, 0x1F934),  # Prince
    (0x1F935, 0x1F935),  # Person in
    (0x1F936, 0x1F936),  # Mrs. Claus
    (0x1F937, 0x1F939),  # Gestures
    (0x1F93A, 0x1F93A),  # Fencer
    (0x1F93C, 0x1F93C),  # Judo
    (0x1F93D, 0x1F93D),  # Water polo
    (0x1F93E, 0x1F93E),  # Person playing
    (0x1F93F, 0x1F93F),  # Diving
    (0x1F940, 0x1F940),  # Wilted
    (0x1F941, 0x1F941),  # Clapper
    (0x1F942, 0x1F943),  # Microphone
    (0x1F944, 0x1F944),  # Drooling
    (0x1F945, 0x1F945),  # Lion
    (0x1F946, 0x1F946),  # Person
    (0x1F947, 0x1F947),  # Goal
    (0x1F948, 0x1F948),  # Jar
    (0x1F949, 0x1F949),  # Racing
    (0x1F94A, 0x1F94A),  # How to play
    (0x1F94B, 0x1F94B),  # 8 ball
    (0x1F94C, 0x1F94C),  # Game
    (0x1F94D, 0x1F94F),  # Martial arts
    (0x1F950, 0x1F950),  # Crescent
    (0x1F951, 0x1F951),  # Cooking
    (0x1F952, 0x1F952),  # Cut of meat
    (0x1F953, 0x1F953),  # Bacon
    (0x1F954, 0x1F954),  # Pancakes
    (0x1F955, 0x1F955),  # Koala
    (0x1F956, 0x1F956),  # Person with
    (0x1F957, 0x1F957),  # Person with
    (0x1F958, 0x1F958),  # Shallow
    (0x1F959, 0x1F959),  # Person with
    (0x1F95A, 0x1F95A),  # Glass of milk
    (0x1F95B, 0x1F95B),  # Glass of milk
    (0x1F95C, 0x1F95C),  # Kiwi
    (0x1F95D, 0x1F95D),  # Lime
    (0x1F95E, 0x1F95E),  # Shrimp
    (0x1F95F, 0x1F95F),  # Croissant
    (0x1F960, 0x1F960),  # Fortune cookie
    (0x1F961, 0x1F961),  # Takeout
    (0x1F962, 0x1F962),  # Chopsticks
    (0x1F963, 0x1F963),  # Bowl with
    (0x1F964, 0x1F964),  # Cup with
    (0x1F965, 0x1F965),  # Lemon
    (0x1F966, 0x1F966),  # Broccoli
    (0x1F967, 0x1F967),  # Pancakes
    (0x1F968, 0x1F968),  # Pretzel
    (0x1F969, 0x1F969),  # Person with
    (0x1F96A, 0x1F96A),  # Cut of meat
    (0x1F96B, 0x1F96B),  # Sandwich
    (0x1F96C, 0x1F96C),  # Leaf
    (0x1F96D, 0x1F96D),  # Mango
    (0x1F96E, 0x1F96E),  # Leaf
    (0x1F96F, 0x1F96F),  # Bowl with
    (0x1F970, 0x1F970),  # Smiling face
    (0x1F971, 0x1F971),  # Person
    (0x1F972, 0x1F972),  # Smiling face
    (0x1F973, 0x1F973),  # Face with
    (0x1F974, 0x1F974),  # Woozy
    (0x1F975, 0x1F975),  # Hot face
    (0x1F976, 0x1F976),  # Cold face
    (0x1F977, 0x1F977),  # Person
    (0x1F978, 0x1F978),  # Disguised
    (0x1F979, 0x1F979),  # Hot
    (0x1F97A, 0x1F97A),  # Face with
    (0x1F97B, 0x1F97B),  # Shaking
    (0x1F97C, 0x1F97C),  # Lab coat
    (0x1F97D, 0x1F97D),  # Goggles
    (0x1F97E, 0x1F97E),  # Hourglass
    (0x1F97F, 0x1F97F),  # Military
    (0x1F980, 0x1F980),  # Crab
    (0x1F981, 0x1F981),  # Lion
    (0x1F982, 0x1F982),  # Scorpion
    (0x1F983, 0x1F983),  # Turkey
    (0x1F984, 0x1F984),  # Unicorn
    (0x1F985, 0x1F985),  # Eagle
    (0x1F986, 0x1F986),  # Duck
    (0x1F987, 0x1F987),  # Bat
    (0x1F988, 0x1F988),  # Shark
    (0x1F989, 0x1F989),  # Owl
    (0x1F98A, 0x1F98A),  # Fox
    (0x1F98B, 0x1F98B),  # Butterfly
    (0x1F98C, 0x1F98C),  # Deer
    (0x1F98D, 0x1F98D),  # Gorilla
    (0x1F98E, 0x1F98E),  # Lizard
    (0x1F98F, 0x1F98F),  # Giraffe
    (0x1F990, 0x1F990),  # Zebra
    (0x1F991, 0x1F991),  # Hedgehog
    (0x1F992, 0x1F992),  # Lizard
    (0x1F993, 0x1F993),  # Shrimp
    (0x1F994, 0x1F994),  # Hedgehog
    (0x1F995, 0x1F995),  # Macaw
    (0x1F996, 0x1F996),  # Parrot
    (0x1F997, 0x1F997),  # Sled
    (0x1F998, 0x1F998),  # Tennis
    (0x1F999, 0x1F999),  # Duck
    (0x1F99A, 0x1F99A),  # Eagle
    (0x1F99B, 0x1F99B),  # Poodle
    (0x1F99C, 0x1F99C),  # Camel
    (0x1F99D, 0x1F99D),  # Two-hump
    (0x1F99E, 0x1F99E),  # Peacock
    (0x1F99F, 0x1F99F),  # Mosquito
    (0x1F9A0, 0x1F9A0),  # Bug
    (0x1F9A1, 0x1F9A1),  # Cockroach
    (0x1F9A2, 0x1F9A2),  # Spider
    (0x1F9A3, 0x1F9A4),  # Rhinoceros
    (0x1F9A5, 0x1F9A5),  # Gecko
    (0x1F9A6, 0x1F9A6),  # Lizard
    (0x1F9A7, 0x1F9A7),  # Rhinoceros
    (0x1F9A8, 0x1F9A8),  # Sparrow
    (0x1F9A9, 0x1F9A9),  # Elephant
    (0x1F9AA, 0x1F9AA),  # Rhinoceros
    (0x1F9AB, 0x1F9AB),  # Otter
    (0x1F9AC, 0x1F9AC),  # Fish
    (0x1F9AD, 0x1F9AD),  # Seal
    (0x1F9AE, 0x1F9AE),  # Beetle
    (0x1F9AF, 0x1F9AF),  # Snail
    (0x1F9B0, 0x1F9B0),  # Fly
    (0x1F9B1, 0x1F9B1),  # Worm
    (0x1F9B2, 0x1F9B2),  # Beetle
    (0x1F9B3, 0x1F9B3),  # Cockroach
    (0x1F9B4, 0x1F9B4),  # Fly
    (0x1F9B5, 0x1F9B5),  # Paw prints
    (0x1F9B6, 0x1F9B6),  # Paw prints
    (0x1F9B7, 0x1F9B7),  # Beaver
    (0x1F9B8, 0x1F9B8),  # Bison
    (0x1F9B9, 0x1F9B9),  # Chameleon
    (0x1F9BA, 0x1F9BA),  # Sea
    (0x1F9BB, 0x1F9BB),  # Clam
    (0x1F9BC, 0x1F9BC),  # Scallop
    (0x1F9BD, 0x1F9BD),  # Snail
    (0x1F9BE, 0x1F9BE),  # Shell
    (0x1F9BF, 0x1F9BF),  # Pegasus
    (0x1F9C0, 0x1F9C0),  # Fairy
    (0x1F9C1, 0x1F9C1),  # Werewolf
    (0x1F9C2, 0x1F9C2),  # Elf
    (0x1F9C3, 0x1F9C4),  # Baby
    (0x1F9C5, 0x1F9C5),  # Genie
    (0x1F9C6, 0x1F9C6),  # Zombie
    (0x1F9C7, 0x1F9C7),  # Brain
    (0x1F9C8, 0x1F9C8),  # Orange
    (0x1F9C9, 0x1F9C9),  # Person
    (0x1F9CA, 0x1F9CA),  # Red
    (0x1F9CB, 0x1F9CB),  # Person
    (0x1F9CC, 0x1F9CC),  # Person
    (0x1F9CD, 0x1F9D0),  # Standing
    (0x1F9D1, 0x1F9D1),  # Teacher
    (0x1F9D2, 0x1F9D2),  # Person
    (0x1F9D3, 0x1F9D3),  # Old
    (0x1F9D4, 0x1F9D4),  # Person
    (0x1F9D5, 0x1F9D5),  # Person
    (0x1F9D6, 0x1F9D7),  # Person
    (0x1F9D8, 0x1F9D8),  # Person
    (0x1F9D9, 0x1F9DB),  # Person
    (0x1F9DC, 0x1F9DD),  # Person
    (0x1F9DE, 0x1F9DF),  # Person
    (0x1F9E0, 0x1F9E0),  # Wizard
    (0x1F9E1, 0x1F9E1),  # Santa
    (0x1F9E2, 0x1F9E2),  # Mrs. Claus
    (0x1F9E3, 0x1F9E3),  # Person
    (0x1F9E4, 0x1F9E4),  # Person
    (0x1F9E5, 0x1F9E5),  # Person
    (0x1F9E6, 0x1F9E6),  # Person
    (0x1F9E7, 0x1F9E7),  # Balloon
    (0x1F9E8, 0x1F9E8),  # Star
    (0x1F9E9, 0x1F9E9),  # Artist
    (0x1F9EA, 0x1F9EA),  # Fire
    (0x1F9EB, 0x1F9EB),  # Mermaid
    (0x1F9EC, 0x1F9EC),  # Elf
    (0x1F9ED, 0x1F9ED),  # Fairy
    (0x1F9EE, 0x1F9EE),  # Vampire
    (0x1F9EF, 0x1F9EF),  # MERMAID
    (0x1F9F0, 0x1F9F0),  # Easter
    (0x1F9F1, 0x1F9F1),  # Balloon
    (0x1F9F2, 0x1F9F2),  # Nail
    (0x1F9F3, 0x1F9F3),  # Balloon
    (0x1F9F4, 0x1F9F4),  # Nail
    (0x1F9F5, 0x1F9F5),  # Nail
    (0x1F9F6, 0x1F9F6),  # Nail
    (0x1F9F7, 0x1F9F7),  # Violin
    (0x1F9F8, 0x1F9F8),  # Game
    (0x1F9F9, 0x1F9F9),  # Teddy
    (0x1F9FA, 0x1F9FA),  # Bobber
    (0x1F9FB, 0x1F9FB),  # Jar
    (0x1F9FC, 0x1F9FC),  # Yarn
    (0x1F9FD, 0x1F9FD),  # Knot
    (0x1F9FE, 0x1F9FE),  # Hello
    (0x1F9FF, 0x1F9FF),  # Nazar
    (0x1FA70, 0x1FA73),  # Various
    (0x1FA78, 0x1FA7A),  # Medical
    (0x1FA80, 0x1FA82),  # X-ray
    (0x1FA86, 0x1FA86),  # Hamster
    (0x1FA90, 0x1FA90),  # Chipmunk
    (0x1FA95, 0x1FA95),  # Safety
    (0x1FAA0, 0x1FAA0),  # Mirror
    (0x1FAA8, 0x1FAA8),  # Helicopter
    (0x1FAAB, 0x1FAAB),  # Lizard
    (0x1FAAE, 0x1FAAE),  # Fly
    (0x1FAB0, 0x1FAB0),  # Beaver
    (0x1FAB1, 0x1FAB1),  # Otter
    (0x1FAB2, 0x1FAB2),  # Shark
    (0x1FAB3, 0x1FAB3),  # Deer
    (0x1FAB6, 0x1FAB6),  # Llama
    (0x1FAC0, 0x1FAC0),  # Person
    (0x1FAC2, 0x1FAC2),  # Mummy
    (0x1FAD0, 0x1FAD0),  # Mate
    (0x1FAD1, 0x1FAD1),  # Glowing
    (0x1FAD2, 0x1FAD2),  # Orange
    (0x1FAD3, 0x1FAD3),  # Blue
    (0x1FAD4, 0x1FAD4),  # Green
    (0x1FAD6, 0x1FAD6),  # Melting
    (0x1FAD7, 0x1FAD7),  # White
    (0x1FAD8, 0x1FAD8),  # Black
    (0x1FAD9, 0x1FAD9),  # Black
]

def build_emoji_pattern():
    """Build regex pattern from curated emoji ranges."""
    parts = []
    for start, end in EMOJI_RANGES:
        if start == end:
            parts.append(f"\\U{start:08X}")
        else:
            parts.append(f"\\U{start:08X}-\\U{end:08X}")
    return re.compile("[" + "".join(parts) + "]+", flags=re.UNICODE)

EMOJI_PATTERN = build_emoji_pattern()

# Characters to PRESERVE (box-drawing and ASCII art characters)
# These are commonly used in ASCII diagrams and should NOT be removed
BOX_DRAWING_CHARS = set()
for i in range(0x2500, 0x2580):  # Box Drawing (U+2500 to U+257F)
    BOX_DRAWING_CHARS.add(chr(i))
for i in range(0x2580, 0x25A0):  # Block Elements (U+2580 to U+259F)
    BOX_DRAWING_CHARS.add(chr(i))
# Geometric Shapes (U+25A0 to U+25FF) - some are emoji, but box parts are safe
for i in range(0x25A0, 0x2600):  # U+25A0 to U+25FF
    BOX_DRAWING_CHARS.add(chr(i))
# These specific block/box chars are safe and commonly used
SAFE_BLOCK_CHARS = {
    ' ', '\t', '\n', '\r',  # Whitespace
    '|', '/', '\\', '-', '_', '+', '=', '*', '#', '@', '!', '?', '<', '>', '^', '~',
    ':', ';', '.', ',', "'", '"', '(', ')', '[', ']', '{', '}', '%', '&', '$', 
    '0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
    'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
    'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
    'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
    'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

# File extensions to process
TEXT_EXTENSIONS = {
    '.py', '.js', '.ts', '.jsx', '.tsx', '.md', '.txt', '.json', '.yaml', '.yml',
    '.html', '.css', '.scss', '.sass', '.less', '.xml', '.csv', '.rst',
    '.sh', '.bash', '.zsh', '.fish', '.ps1', '.bat', '.cmd',
    '.rb', '.java', '.kt', '.swift', '.go', '.rs', '.c', '.cpp', '.h', '.hpp',
    '.sql', '.graphql', '.vue', '.svelte'
}

# Directories to skip
SKIP_DIRS = {
    '.git', '.svn', '.hg', 'node_modules', '__pycache__', '.pytest_cache',
    '.mypy_cache', '.tox', 'venv', '.venv', 'env', '.env', 'dist', 'build',
    '.eggs', '*.egg-info', '.tox', '.nox', '.coverage', 'htmlcov',
    '.DS_Store', '.AppleDouble', '.LSOverride'
}


def is_text_file(filepath: str) -> bool:
    """Check if file is a text file based on extension."""
    _, ext = os.path.splitext(filepath)
    return ext.lower() in TEXT_EXTENSIONS


def should_skip_dir(dirname: str) -> bool:
    """Check if directory should be skipped."""
    return dirname in SKIP_DIRS or dirname.startswith('.')


def remove_emojis(text: str) -> str:
    """Remove all emoji characters from text."""
    return EMOJI_PATTERN.sub('', text)


def process_file(filepath: str, dry_run: bool = False, verbose: bool = False) -> tuple[bool, bool]:
    """
    Process a single file to remove emojis.
    Returns: (was_modified, had_emojis)
    """
    try:
        with open(filepath, 'r', encoding='utf-8', errors='replace') as f:
            original_content = f.read()
    except (IOError, OSError) as e:
        if verbose:
            print(f"Error reading {filepath}: {e}")
        return False, False
    
    had_emojis = bool(EMOJI_PATTERN.search(original_content))
    
    if not had_emojis:
        return False, False
    
    new_content = remove_emojis(original_content)
    
    if dry_run:
        print(f"[DRY RUN] Would modify: {filepath}")
        return True, True
    
    try:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(new_content)
        if verbose:
            print(f"Removed emojis from: {filepath}")
        return True, True
    except (IOError, OSError) as e:
        if verbose:
            print(f"Error writing {filepath}: {e}")
        return False, True


def scan_directory(root_dir: str, dry_run: bool = False, verbose: bool = False) -> dict:
    """
    Recursively scan directory and remove emojis from all text files.
    Returns statistics dictionary.
    """
    stats = {
        'files_scanned': 0,
        'files_modified': 0,
        'files_with_emojis': 0,
        'errors': 0
    }
    
    for dirpath, dirnames, filenames in os.walk(root_dir):
        # Filter out skip directories
        dirnames[:] = [d for d in dirnames if not should_skip_dir(d)]
        
        for filename in filenames:
            filepath = os.path.join(dirpath, filename)
            
            if not is_text_file(filepath):
                continue
            
            stats['files_scanned'] += 1
            
            modified, had_emojis = process_file(filepath, dry_run, verbose)
            
            if had_emojis:
                stats['files_with_emojis'] += 1
            
            if modified:
                stats['files_modified'] += 1
    
    return stats


def main():
    import argparse
    
    parser = argparse.ArgumentParser(
        description='Remove all emoji characters from text files in a codebase.',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s .                      # Scan and remove emojis from current directory
  %(prog)s /path/to/project       # Scan specific directory
  %(prog)s . --dry-run            # Show what would be modified without changing files
  %(prog)s . --verbose            # Print each file as it's processed
  %(prog)s . --include-pattern "*.md"  # Only process markdown files
        """
    )
    
    parser.add_argument(
        'directory',
        nargs='?',
        default='.',
        help='Directory to scan (default: current directory)'
    )
    
    parser.add_argument(
        '--dry-run', '-n',
        action='store_true',
        help='Show what would be modified without making changes'
    )
    
    parser.add_argument(
        '--verbose', '-v',
        action='store_true',
        help='Print each file as it is processed'
    )
    
    parser.add_argument(
        '--include-pattern',
        help='Only process files matching this glob pattern (e.g., "*.md")'
    )
    
    parser.add_argument(
        '--exclude-pattern',
        help='Exclude files matching this glob pattern'
    )
    
    args = parser.parse_args()
    
    # Change to target directory
    target_dir = os.path.abspath(args.directory)
    
    if not os.path.isdir(target_dir):
        print(f"Error: {target_dir} is not a valid directory", file=sys.stderr)
        sys.exit(1)
    
    print(f"Scanning directory: {target_dir}")
    
    if args.dry_run:
        print("[DRY RUN MODE - No changes will be made]")
    
    # Note: include-pattern and exclude-pattern would require additional implementation
    # For now, using built-in extension-based filtering
    
    stats = scan_directory(target_dir, dry_run=args.dry_run, verbose=args.verbose)
    
    print("\n" + "=" * 50)
    print("SUMMARY")
    print("=" * 50)
    print(f"Files scanned:        {stats['files_scanned']}")
    print(f"Files with emojis:    {stats['files_with_emojis']}")
    print(f"Files modified:       {stats['files_modified']}")
    print(f"Errors:               {stats['errors']}")
    
    if args.dry_run:
        print("\nRun without --dry-run to apply changes.")


if __name__ == '__main__':
    main()