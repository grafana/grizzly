---
title: "Cupper Shortcodes"
date: 2019-02-12T23:39:06-06:00
tags: [hugo, shortcodes]
toc: true
---

## blockquote

```
{{</* blockquote author="Carl Jung" */>}}
Even a happy life cannot be without a measure of darkness, and the word happy would lose its meaning if it were not balanced by sadness. It is far better to take things as they come along with patience and equanimity.
{{</* /blockquote */>}}
```

{{< blockquote author="Carl Jung" >}}
Even a happy life cannot be without a measure of darkness, and the word happy would lose its meaning if it were not balanced by sadness. It is far better to take things as they come along with patience and equanimity.
{{< /blockquote >}}

## note

```
{{</* note */>}}
This is a note! It's something the reader may like to know about but is supplementary to the main content. Use notes when something may be interesting but not critical. You can also *include* **markdown** stuffs like `code`. 
{{</* /note */>}}
```

{{< note >}}
This is a note! It's something the reader may like to know about but is supplementary to the main content. Use notes when something may be interesting but not critical. You can also *include* **markdown** stuffs like `code`.
{{< /note >}}

## warning note

```
{{</* warning */>}}
This is a warning! It's about something the reader should be careful to do or to avoid doing. Use warnings when something could go wrong. You can also *include* **markdown** stuffs like `code`.
{{</* /warning */>}}
```

{{< warning >}}
This is a warning! It's about something the reader should be careful to do or to avoid doing. Use warnings when something could go wrong. You can also *include* **markdown** stuffs like `code`.
{{< /warning >}}

## cmd

```
{{</* cmd */>}}
hugo server --gc
{{</* /cmd */>}}
```

{{< cmd >}}
hugo server --gc
{{< /cmd >}}

## code

```
{{</* code numbered="true" */>}}
<div [[[role="dialog"]]] [[[aria-labelledby="dialog-heading"]]]>
  <button [[[aria-label="close"]]]>x</button>
  <h2 [[[id="dialog-heading"]]]>Confirmation</h2>
  <p>Press Okay to confirm or Cancel</p>
  <button>Okay</button>
  <button>Cancel</button>
</div>
{{</* /code */>}}

1. The dialog is only announced as a dialog if it takes the `dialog` ARIA role
2. The `aria-labelledby` relationship attribute makes the element carrying the `id` it points to its label
3. The close button uses `aria-label` to provide the text label "close", overriding the text content
4. The heading is used as the dialog's label. The `aria-labelledby` attribute points to its `id`
```

{{< code numbered="true" >}}
<div [[[role="dialog"]]] [[[aria-labelledby="dialog-heading"]]]>
  <button [[[aria-label="close"]]]>x</button>
  <h2 [[[id="dialog-heading"]]]>Confirmation</h2>
  <p>Press Okay to confirm or Cancel</p>
  <button>Okay</button>
  <button>Cancel</button>
</div>
{{< /code >}}

1. The dialog is only announced as a dialog if it takes the `dialog` ARIA role
2. The `aria-labelledby` relationship attribute makes the element carrying the `id` it points to its label
3. The close button uses `aria-label` to provide the text label "close", overriding the text content
4. The heading is used as the dialog's label. The `aria-labelledby` attribute points to its `id`

## syntax highlighting

To get syntax highlighting for your code, use markdown code fences, then specify the language:

````
```html
<div role="dialog" aria-labelledby="dialog-heading">
  <button aria-label="close">x</button>
  <h2 id="dialog-heading">Confirmation</h2>
  <p>Press Okay to confirm or Cancel</p>
  <button>Okay</button>
  <button>Cancel</button>
</div>
```
````

```html
<div role="dialog" aria-labelledby="dialog-heading">
  <button aria-label="close">x</button>
  <h2 id="dialog-heading">Confirmation</h2>
  <p>Press Okay to confirm or Cancel</p>
  <button>Okay</button>
  <button>Cancel</button>
</div>
```

## codePen

```
{{</* codePen VpVNKW */>}}
```

{{< codePen VpVNKW >}}

## colors

```
{{</* colors "#111111, #cccccc, #ffffff" */>}}
```

{{< colors "#111111, #cccccc, #ffffff" >}}

## expandable

```
{{</* expandable label="A section of dummy text" level="2" */>}}
Here is some markdown including [a link](https://twitter.com/heydonworks). Donec erat est, feugiat a est sed, aliquet pharetra ipsum. Vivamus in arcu leo. Praesent feugiat, purus a molestie ultrices, libero massa iaculis ante, sit amet accumsan leo eros vel ligula.
{{</* /expandable */>}}
```

{{< expandable label="A section of dummy text" level="2" >}}
Here is some markdown including [a link](https://twitter.com/heydonworks). Donec erat est, feugiat a est sed, aliquet pharetra ipsum. Vivamus in arcu leo. Praesent feugiat, purus a molestie ultrices, libero massa iaculis ante, sit amet accumsan leo eros vel ligula.
{{< /expandable >}}

## fileTree

```
{{</* fileTree */>}}
* Level 1 folder
    * Level 2 file
    * Level 2 folder
        * Level 3 file
        * Level 3 folder
            * Level 4 file
        * Level 3 folder
            * Level 4 file
            * Level 4 file
        * Level 3 file
    * Level 2 folder
        * Level 3 file
        * Level 3 file
        * Level 3 file
    * Level 2 file
* Level 1 file
{{</* /fileTree */>}}
```

{{< fileTree >}}
* Level 1 folder
    * Level 2 file
    * Level 2 folder
        * Level 3 file
        * Level 3 folder
            * Level 4 file
        * Level 3 folder
            * Level 4 file
            * Level 4 file
        * Level 3 file
    * Level 2 folder
        * Level 3 file
        * Level 3 file
        * Level 3 file
    * Level 2 file
* Level 1 file
{{< /fileTree >}}

## ticks

```
{{</* ticks */>}}
* Selling point one
* Selling point two
* Selling point three
{{</* /ticks */>}}
```

{{< ticks >}}
* Selling point one
* Selling point two
* Selling point three
{{< /ticks >}}

## figureCupper

```
{{</* figureCupper
img="sun.jpg" 
caption="The Sun is the star at the center of the Solar System. It is a nearly perfect sphere of hot plasma, with internal convective motion that generates a magnetic field via a dynamo process. It is by far the most important source of energy for life on Earth. [Credits](https://images.nasa.gov/details-GSFC_20171208_Archive_e000393.html)." 
command="Resize" 
options="700x" */>}}
```

{{< figureCupper
img="sun.jpg" 
caption="The Sun is the star at the center of the Solar System. It is a nearly perfect sphere of hot plasma, with internal convective motion that generates a magnetic field via a dynamo process. It is by far the most important source of energy for life on Earth. [Credits](https://images.nasa.gov/details-GSFC_20171208_Archive_e000393.html)." 
command="Resize" 
options="700x" >}}

## principles

See the [full principles list](https://github.com/zwbetz-gh/cupper-hugo-theme/blob/master/data/principles.json).

```
{{</* principles include="Add value, Be consistent" descriptions="true" */>}}
```

{{< principles include="Add value, Be consistent" descriptions="true" >}}

## wcag

See the [full wcag list](https://github.com/zwbetz-gh/cupper-hugo-theme/blob/master/data/wcag.json). 

```
{{</* wcag include="1.2.1, 1.3.1, 4.1.2" */>}}
```

{{< wcag include="1.2.1, 1.3.1, 4.1.2" >}}

## tested

See the [full browser list](https://github.com/zwbetz-gh/cupper-hugo-theme/tree/master/static/images).

```
{{</* tested using="Firefox + JAWS, Chrome, Safari iOS + Voiceover, Edge" */>}}
```

{{< tested using="Firefox + JAWS, Chrome, Safari iOS + Voiceover, Edge" >}}
