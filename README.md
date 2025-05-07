# wap-tool

Automate the generation of WAPs (Wochenarbeitsplan).

## Quickstart
TODO: how to obtain the image/script/data

Goal: edit a WAP as yaml and print it.
We create a minimal example:
``` yaml
meta: # define metadata
  author: Autor
  firstDay: 2025-04-20
  startTime: '05:30'
  endTime: '23:00'
  unit: 'Kp 42/3'
  title: Det X
  version: 'Stand: 19.04.2025'
categories:
- identifier: Orange
  color: '#ff9900'
weeks:
- remarks:
  - Hier ist eine Wochenbemerkung.
  days:
  - remarks:
      - Das ist eine Tagesbemerkung
      - Hier ist eine weitere
    columns: # split the day in these columns
    - Det
    - Beso
    events:
    - title: Zi-Bezug & Zi Ordnung
      start: '10:00' # must give a start time
      end: '11:00' # as well as an end time
      appearsIn: # in which columns, defined above, the event appears
      - Det
      category: Orange # references cateogiry defined above
      description: Det C, GK 106 & GK 108
    - title: Einrückungs & Fahr Rape
      start: '10:00'
      end: '11:00'
      appearsIn:
      - Det
      category: Orange
      description: Einh Four  # optionally a description can be given
    - title: Kontrolle AWB Karte
      start: '10:00'
      end: '10:30'
      footnote: true # Footnotes are printed below the day
      appearsIn:
      - Beso
      description: Fkhaus 15, Det C
    - title: Fassung IAA-Mat
      start: '11:00'
      end: '12:00'
      appearsIn:
      - Det
      category: Orange
      description: Einh Fw, Mat Mag
```
We save it to `data/minimal.yaml`.
Running
``` sh
./run-docker.sh data/minimal.yaml examples/minimal.pdf
```
Produces the file `examples/minimal.pdf`.
We see the events we defined ![](examples/snip.png)

You might have obtained a template to get started or can use the folder `data/` for inspiration.

## Docs
The [schema](schema/wap.json) remains the source of truth for the model.
We describe the most important fields.
- `meta` about the WAP.
- `categories` used for styling.
- `weeks`
    - `remarks`
    - `days` 
        - `remarks`
        - `columns`
        - `days`
            - `events`
- TODO footnotes

The [vscode-yaml](https://github.com/redhat-developer/vscode-yaml) extension is recommended.
Autocompletion is available and validations are available.
For example, dates are checked to be in the right format.

Associate the correct schema with your file by either including the following line in your yaml file (you might have to modify the path):
``` yaml
# yaml-language-server: $schema=../schema/wap.json
```
Or, by associating the schema definition with a glob pattern in your `settings.json`:
``` yaml
{
    "yaml.schemas": {
       "./schema/wap.json": "*"
    },
}
```
## Project
### Background
Most WKs share the similar format.
By editing a template, a new WAP should be easily defined.

Currently, WAPs are mostly edited in Excel.
Layouting and formating is done manually, events are simply shapes overlayed over the cells.
However, Excel provides great flexibility and is widely known.

Miloffice provides similar capabilities, but it is slow and has a bad user experience.

WAPs have a typical format: each page show a week. Days are columns, that can be further subdivided in subcolumns.
Events are drawn in the week, have a description and are styled in a certain way.
Additionaly, remarks can be added for the day or the entire week.

While we tried to reproduce this original format, **the WAP format is not regulated and could be changed**.

While we target the use in our company first, it could be reused in other places as this is a general problem in the military.
<!-- specialities: not a typical calendar -->

### Implementation
- Intermediate Data Format: define the WAP data
    - model defined as a [JSON schema](https://json-schema.org/)
    - edited as YAML
    - A human and machine friendly format, that is widely known and comes with good tool support
    - We tried to make it easy to adapt an existing WAP. For example, to change the date we must only change the initial date. All other days are offset from this date.
- Backend: print the WAP
    - at the moment, a Go tool can generate pdfs


### Future Work
1. Create a frontend to simplify editing
    - editing the yaml by hand can be tedious as well: For example, resizing an event is easier than edit timestamps or dragging events instead of copy-and-pasting lines.
    - there are many pitfalls, and it could be challenging to support all features
    - import/export as yaml
    - could reuse/reimlement the existing layout algo

2. Alternatively, provide tool to simplify editing the yaml.
    - Shift all start/end times for a day after a given time
    - Harmonize the data: sort the events per day, use the same order of fields, ...
    - Delete a column (for a day)

3. Generate *Tagesbefehle*. Use the same or similar format to generate detailed plans for a single day.

## Development
For the Go backend, see [go/README.md](./go/README.md)