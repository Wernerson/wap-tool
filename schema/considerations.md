Data model for a WAP
# Key points
- Defined as a [JSON schema](https://json-schema.org/) 
- Written in YAML
- To be used as an intermediary data format
    - possibly to store and use as templates
    - import/export from a frontend
    - used as the input for printing backends

# Entities
Check the spec directly (TODO link)

<!-- TODO Include examples -->
- `Metadata` about the WAP.
- `Category` used for styling.
- `Day` holding events and have columns.
- `Event` has a start and end time.  Locations and responsible persons can be optionally assigned.
- `Remarks` can be given, to be displayed as additional information on the side.

Per event columns can be defined in which it should appear.
Alternatively, we can use an optional field `footnote` on an event.
Footnotes are events in the "Beso" column that are fully described in the part below the WAP.

# Tooling
The [vscode-yaml](https://github.com/redhat-developer/vscode-yaml) extension is recommended.
Autocompletion is available and validations are available.
For example, dates are check to be in the right format.

Associate the correct schema with your file by either including the following line in your yaml file (you might have to modify the path):
``` yaml
# yaml-language-server: $schema=../schema/wap.json
```
Or, by associating the schema definition with a glob pattern in your `settings.json`:
``` yaml
{
    "yaml.schemas": {
       "./data/schema_v2.json": "v2_*"
    },
}
```
# Implementation decisions
- Use yaml, a human and machine friendly format, that is widely known and comes with good tool support (alternatives: TOML, JSON)
- Days are relative with an offset relative to an initial start date. This way it is easier to use a template; only a single element needs to be modified.
- Dates are included as ISO8601 strings
- In the examples, days can have different columns, so this is not globally defined.
- In some cases, events are split within a column

# References
- [json-schema reference](https://json-schema.org/reference)
- [yaml spec](https://yaml.org/spec/1.2.2/)

