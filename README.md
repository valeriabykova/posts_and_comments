# API системы постов и комментариев 

## Задача
Реализовать систему для добавления и чтения постов и комментариев с использованием GraphQL, аналогичную комментариям к постам на популярных платформах, таких как Хабр или Reddit.

Для реализации использовались библиотеки `gqlgen` и `gorm`.

## Характеристики системы постов:
- Можно просмотреть список постов.
В примере – три первых поста на странице, каждый – с одним комментарием.

```graphql
query GetPosts {
  posts(page: 1, limit: 3) {
    id
    title
    author
    comments(first: 1) {
      edges {
        node {
          body
        }
      }
    }
  }
}

```
- Можно просмотреть пост и комментарии под ним.
Здесь можно получить запись по ее id, а также первые 3 комментария + по одному ответу к каждому из комментариев.
```graphql
query GetPost {
  post(id: 1) {
    title
    body
    author
    comments(first: 3) {
      edges {
        node {
          id
          body
          replies(first: 1) {
            edges {
              node {
                id
                body
              }
            }
          }
        }
      }
    }
  }
}

```

- Пользователь, написавший пост, может запретить оставление комментариев к своему посту. Для этого в структуре существует поле `allowComments`

Для выполнения запросов необходимо перейти на Playground от GraphQL (создается при запуске программы), и выполнить запросы сначала для созданию постов/комментариев, а после – для  их вывода
- Создание поста:
```graphql
mutation NewPost {
   createPost(post: {title:"lorem", body:"ipsum", author: "me", allowComments: true}) {
     id
     body
   }
 }

```
- Создание комментария к посту
```graphql
mutation NewCommentToPost {
  createComment(comment: {postId: 1, body:"it's true", author:"anonym"}) {
    id
  }
}
```
- Создание ответа на комментарий

```graphql
mutation NewCommentToComment {
  createComment(comment: {postId: 1, parentId: 1, body:"no u", author:"me"}) {
    id
    parentId
  }
}
```

## Характеристики системы комментариев к постам:
- Комментарии организованы иерархически, позволяя вложенность без ограничений.
- Длина текста комментария ограничена до, например, 2000 символов.
- Система пагинации для получения списка комментариев.
    - Пагинация для комментариев выполнена при помощи курсоров, пагинация для постов выполнена как LIMIT-OFFSET.

## Требования к реализации:
- Система должна быть написана на языке Go.
- Использование Docker для распространения сервиса в виде Docker-образа.
- Хранение данных может быть как в памяти (in-memory), так и в PostgreSQL. Для этого в Dockerfile прописаны переменные окружения, в которые можно передать либо `Postgres` либо `In-memory` для выбора. По умолчанию используется In-memory вариант. 
- Покрытие реализованного функционала unit-тестами. (К сожалению, не успела.) 

Для запуска работы приложения необходимо выполнить команды docker-compose:
```
docker compose build
docker compose up
```