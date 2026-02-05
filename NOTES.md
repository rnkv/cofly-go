Object is considered as array patch only if its keys are following the pattern:
`<indexFrom>..[<indexTo>]: <value>`

where:
- `<indexFrom>` is the number of the index of the first element to be modified
- `<indexTo>` is the number of the index of the last element to be modified (not inclusive), if not provided, it is considered the same as `<indexFrom>`
- `<value>` is the value to be modified (array)

Example:
```
{
	"-1..0": [1, 2, 3],
    "1..3": [4, 5, 6],
    "3..": [7, 8, 9],
    "5..5": [10],
}
```