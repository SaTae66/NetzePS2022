[![Go](https://github.com/SaTae66/NetzePS2022/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/SaTae66/NetzePS2022/actions/workflows/go.yml)
# NetzePS2022
Erstelle ein Transmit- (TX) und Receive-(RX) Programm in jeweils zwei verschiedenen Programmiersprachen, das mittels UDP eine Datei zwischen den vier Kombinationen schnellstens und fehlerfrei übertragen kann.

V1:
In der ersten Version sollen KEINE Kontrollnachrichten zwischen TX und RX verwendet werden!
Die Fehlerfreiheit soll nur mittels Prüfsumme oder Hash sichergestellt werden.

V2:
In der zweiten Version sollen die Programme um ein Stop&Wait Protokoll erweitert werden.
D.h. der Empfänger bestätigt den Erhalt der Nachricht, bevor der Sender weitersendet.
