---
title: "Importing INE Shapefiles into a Spatialite Database"
date: 2024-07-14T13:00:00-07:00
draft: false
author: "Hector Yeomans"
description: "Step-by-step guide to importing Mexico's National Electoral Institute shapefiles into a Spatialite database using macOS and QGIS."
tags: ["gis", "spatialite", "qgis", "mexico", "INE", "shapefiles", "mapping"]

showToc: true
TocOpen: false
hidemeta: false

disableShare: false
disableHLJS: false
hideSummary: false
searchHidden: false
ShowReadingTime: true
ShowBreadCrumbs: true
cover:
  image: "/img/shapefile_qgis.gif"
  alt: "import_shape_file_spatialite"
  caption: "Importing layer into QGIS"
  relative: false # when using page bundles set this to true
  hidden: false # only hide on current single page
---

In this post, we'll walk through the process of importing Mexico's National Electoral Institute (INE) shapefiles into a Spatialite database. We'll be using macOS and QGIS for this tutorial.
Setting Up Spatialite

First, let's install Spatialite using Homebrew:

```sh
brew install spatialite-tools
```

Once installed, create a new Spatialite database by running:

```sh
spatialite ine.db
```

You should see output indicating the Spatialite version and supported extensions.

```sql
SpatiaLite version ..: 5.1.0    Supported Extensions:
        - 'VirtualShape'        [direct Shapefile access]
        - 'VirtualDbf'          [direct DBF access]
        - 'VirtualText'         [direct CSV/TXT access]
        - 'VirtualGeoJSON'              [direct GeoJSON access]
        - 'VirtualXL'           [direct XLS access]
        - 'VirtualNetwork'      [Dijkstra shortest path - obsolete]
        - 'RTree'               [Spatial Index - R*Tree]
        - 'MbrCache'            [Spatial Index - MBR cache]
        - 'VirtualFDO'          [FDO-OGR interoperability]
        - 'VirtualBBox'         [BoundingBox tables]
        - 'VirtualSpatialIndex' [R*Tree metahandler]
        - 'VirtualElementary'   [ElemGeoms metahandler]
        - 'VirtualRouting'      [Dijkstra shortest path - advanced]
        - 'VirtualKNN2' [K-Nearest Neighbors metahandler]
        - 'VirtualGPKG' [OGC GeoPackage interoperability]
        - 'VirtualXPath'        [XML Path Language - XPath]
        - 'SpatiaLite'          [Spatial SQL - OGC]
PROJ version ........: Rel. 9.4.0, March 1st, 2024
GEOS version ........: 3.12.1-CAPI-1.18.1
RTTOPO version ......: 1.1.0
TARGET CPU ..........: x86_64-apple-darwin23.0.0
the SPATIAL_REF_SYS table already contains some row(s)
SQLite version ......: 3.45.3
Enter ".help" for instructions
SQLite version 3.45.3 2024-04-15 13:34:05
Enter ".help" for instructions
Enter SQL statements terminated with a ";"
```

## Downloading INE Shapefiles

- Visit the INE transparency portal: https://pautas.ine.mx/transparencia/mapas/

![](/img/pautas_ine.jpg)

- Download the shapefiles you need. For this tutorial, we'll use the ENTIDADES (States) shapefile.

## Importing Shapefiles into QGIS

1. Open QGIS and drag and drop the downloaded shapefile into the QGIS window.
   ![](/img/shapefile_qgis.gif)
2. Establish a connection to your Spatialite database (ine.db) in QGIS.
   ![](/img/ine_db_conn.jpg)
3. Import the shapefile layer into the Spatialite database using QGIS's import functionality.
   ![](/img/entidad_table.jpg)

## Querying the Database

Now that we've imported the data, we can query it using Spatialite. Connect to your database in the terminal:

```sh
spatialite ine.db
```

Here's an example query to check if a specific latitude and longitude are within a state's geometry:

```
SELECT entidad,
       nombre,
       circunscri
FROM ENTIDAD
WHERE ST_Contains(geom, MakePoint(-110.969966939491, 29.142818450243844));
```

This query will return the state information for the given coordinates.

![](/img/output_entidad.jpg)

## Conclusion

You've now successfully imported INE shapefiles into a Spatialite database and can perform spatial queries on the data. This setup allows for efficient spatial data management and analysis of Mexican electoral geography.

Remember to adjust the coordinates in the sample query to match the specific location you're interested in analyzing.
