package products

import (
	"context"
	"database/sql"
	"server-api-admin/models"
)

func FetchSpecs(ctx context.Context, tx *sql.Tx) (
	[]models.Specification,
	[]models.Specification,
	map[int]models.ProductMainType,
	error,
) {
	materials := make([]models.Specification, 0)
	metalColors := make([]models.Specification, 0)
	productTypes := make(map[int]models.ProductMainType)

	// materials
	rows, err := tx.QueryContext(
		ctx,
		`
			SELECT material_id, name
			FROM product_material
			ORDER BY material_id ASC
		`,
	)
	if err != nil {
		return materials, metalColors, productTypes, err
	}

	defer rows.Close()

	for rows.Next() {
		var s models.Specification
		err = rows.Scan(&s.ID, &s.Name)
		if err != nil {
			return materials, metalColors, productTypes, err
		}
		materials = append(materials, s)
	}

	// metal colors
	rows, err = tx.QueryContext(
		ctx,
		`
			SELECT color_id, name
			FROM metal_color
			ORDER BY color_id ASC
		`,
	)
	if err != nil {
		return materials, metalColors, productTypes, err
	}

	defer rows.Close()

	for rows.Next() {
		var s models.Specification
		err = rows.Scan(&s.ID, &s.Name)
		if err != nil {
			return materials, metalColors, productTypes, err
		}
		metalColors = append(metalColors, s)
	}

	// product types
	rows, err = tx.QueryContext(
		ctx,
		`
			SELECT 
				pt.product_type_id, 
				pt.main_type_id,
				pmt.name,
				pst.name
			FROM product_type pt
			JOIN product_main_type pmt ON pmt.main_type_id = pt.main_type_id
			LEFT JOIN product_sub_type pst ON pst.sub_type_id = pt.sub_type_id
		`,
	)
	if err != nil {
		return materials, metalColors, productTypes, err
	}

	defer rows.Close()

	for rows.Next() {
		var productTypeID, mainTypeID int
		var mainTypeName string
		var subTypeName sql.NullString

		err = rows.Scan(&productTypeID, &mainTypeID, &mainTypeName, &subTypeName)
		if err != nil {
			return materials, metalColors, productTypes, err
		}

		if _, ok := productTypes[mainTypeID]; !ok {
			productTypes[mainTypeID] = models.ProductMainType{
				Name:     mainTypeName,
				Subtypes: make([]models.Specification, 0),
			}
		}

		if subTypeName.Valid {
			tempArr := append(
				productTypes[mainTypeID].Subtypes,
				models.Specification{
					ID:   productTypeID,
					Name: subTypeName.String,
				},
			)
			productTypes[mainTypeID] = models.ProductMainType{
				Name:     mainTypeName,
				Subtypes: tempArr,
			}
		} else {
			tempArr := append(
				productTypes[mainTypeID].Subtypes,
				models.Specification{
					ID:   productTypeID,
					Name: "(No Subtype)",
				},
			)
			productTypes[mainTypeID] = models.ProductMainType{
				Name:     mainTypeName,
				Subtypes: tempArr,
			}
		}
	}

	return materials, metalColors, productTypes, nil
}
