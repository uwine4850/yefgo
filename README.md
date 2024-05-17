Для роботи дана бібліотека потребує конфігураційний файл, який генерую бібліотека [yefpy](https://github.com/uwine4850/yefpy).

Цей проект створений для того, щоб запускати код написаний на Python всередині кода Golang.
Приклад роботи цієї бібліотеки показаний у наступному проекті https://github.com/uwine4850/yefexample.

## Початок роботи
Глобально роботи із цією бібліотекую розділена на дві частини, а саме на генерацію коду та використання потрібних об'єктів.
Також для кожної сесії консолі потрібно запускати файл `start.sh`.

### Генерація коду
Для того, щоб здійснити генерацію коду потрібен конфігураційний файл створений за домомогою [yefpy](https://github.com/uwine4850/yefpy).
Після того як файл .yaml створений потрібно згенерувати код golang наступним чином:
```
func main() {
    err := codegen.Generate("pygen/pygen.yaml", "github.com/uwine4850/yefgotest")
    if err != nil {
    	panic(err)
    }
}
```
У функцію `Generate` передається шлях до конфігурації та назва модуля проекту, який можна занайти у файлі `go.mod`.
Код буде згенеровано у директорію `gen`.

### Використання згенерованого коду
Даний проект запускає код Python, він не переписує його на мову Golang, тому для правильної роботи потрібен установлений 
інтерпритатор Python. Також проект Python повинен бути на локальній машині.

Щоб почати працювати із Python його завжди потрібно ініціалізувати наступним чином:
```
init := module.InitPython{}
init.Initialize()
defer init.Finalize()
```
Цей код просто ініціалізує Python, щоб мова Golang мала доступ до нього.

Далі потрібно отримати доступ до модулів Python. Важливо зазначити, що доступ потрібен саме до модулів проекту який 
використовується, тому шлях імпотрту повинен бути відповідний.
```
shopMod, err := init.GetPyModule("proj.shop")
if err != nil {
	panic(err)
}
customerMod, err := init.GetPyModule("proj.customer")
if err != nil {
	panic(err)
}
```
Завдяки цим методам тепер проект має доступ до потрібних модулів.

Після цього можна використовувати згенерований код. Приклад використання:
```
shopClass, err := shop.Shop{}.New(&init, shopMod)
if err != nil {
	panic(err)
}
productClass, err := shop.Product{}.New(&init, shopMod, "prod1")
if err != nil {
	panic(err)
}
customerClass, err := customer.Customer{}.New(&init, customerMod, &shopClass.Class)
if err != nil {
	panic(err)
}
err = shopClass.Add_product(&productClass.Class)
if err != nil {
	panic(err)
}
fmt.Println(shopClass.Get_products())
err = customerClass.By_product(&productClass.Class)
if err != nil {
	panic(err)
}
fmt.Println(shopClass.Get_products())
```
Кожен метод, який тут показаний викликає метод напряму із Python. Golang ніяк не взаємодіє із тілом метода, тому воно 
може вільно змінюватися.

### Загальний алгоритм роботи
- Створити конфігураційний файл та файл start.sh за допомогою [yefpy](https://github.com/uwine4850/yefpy).
- Запустити `start.sh` для кожної сесії консолі де вткористовується даний проект. Запустити можна так `. start.sh`.
- Згенерувати код Golang.
- Ініціалізувати Python.
- Запустити потрібний функціонал./home/fhx/GolandProjects/yefgotest