# Pruebas

| Test | Regla |
|---|---|
| [TestPost_UpdateOnlyListedFields](../tests/post_test.go) | actualizar `["excerpt"]` cambia solo el extracto; el título queda igual |
| [TestPost_AuthorCannotApprove](../tests/post_test.go) | un autor que actualiza `state` a Approved recibe un error de campo en `state` y el artículo sigue en borrador |
| [TestCreate_PublishesChange](../tests/post_test.go) | crear un artículo publica un `*crud.Change` en `crud.Topic(ModelName, "Post", model.Create)` |
| [TestTransition_ApproveFromInReview](../tests/transition_test.go) | una persona revisora aprueba un artículo en revisión: el artículo queda Approved y la transición guarda quién y desde qué estado |
| [TestTransition_ChangesRequestedNeedsReason](../tests/transition_test.go) | pedir cambios sin motivo es un error de campo en `reason` |
| [TestTransition_NoUpdateOrDelete](../tests/transition_test.go) | actualizar o borrar una transición no existe: `unknown operation` |
| [TestTenant_ReadOtherTenantIsNotFound](../tests/tenant_test.go) | un artículo del inquilino a leído por el inquilino b es `model.ErrorNotFound` |
| [TestMount_MatchesDefinition](../tests/tenant_test.go) | lo que monta `MountOperations` coincide con `Definition()` |
| [TestDemo_Valid](../tests/demo_test.go) | `seed.Demo()` (wasm) es válida, su zona carga y entrega al menos una pantalla |
| [TestHappyPathLifecycle](../tests/editorial_test.go) | el ciclo borrador, revisión, aprobado y publicado recorre el módulo completo |
| [TestSentinelTexts](../tests/editorial_test.go) | los errores centinela conservan su texto |
| [TestViewPresenter](../tests/view_test.go) | el presentador de artículos se construye con el título y el buscador |
| [TestItemsDeclaredOnFields](../tests/view_test.go) | los campos de etiqueta, descripción e insignia están declarados en el modelo |
| [TestMigrate_UpToDate_OnlyEnsuresTables](../tests/migrate_test.go) | con el esquema al día solo se aseguran las tablas |
| [TestMigrate_NewField_AddsOnlyThatColumn](../tests/migrate_test.go) | un campo nuevo agrega solo esa columna |
| [TestMigrate_IntrospectionFails_ReturnsErrorAndStops](../tests/migrate_test.go) | si falla la introspección, la migración devuelve el error y se detiene |
