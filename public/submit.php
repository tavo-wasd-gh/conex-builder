<?php
if ($_SERVER["REQUEST_METHOD"] == "POST") {
    $date = new DateTime();
    $isoDate = $date->format(DateTime::ISO8601);

    $directory = "submit/";
    $filename = "$isoDate.md";
    $filePath = $directory . $filename;

    if (!is_dir($directory)) {
        mkdir($directory, 0755, true);
    }

    $title = isset($_POST["title"]) ? $_POST["title"] : "";
    $slogan = isset($_POST["slogan"]) ? $_POST["slogan"] : "";
    $name = isset($_POST["name"]) ? $_POST["name"] : "";
    $email = isset($_POST["email"]) ? $_POST["email"] : "";
    $phone = isset($_POST["phone"]) ? $_POST["phone"] : "";

    $editorContent = isset($_POST["editor"]) ? $_POST["editor"] : "";
    $editorContent = str_replace("\r\n", "\n", $editorContent); // Convert CRLF to LF
    $editorContent = str_replace("\r", "\n", $editorContent); // Convert CR to LF

    $content = <<<EOD
---
date: $isoDate
author: "$name <$email> <$phone>"
title: "$title"
description: "$slogan"
layout: single
---

$editorContent
EOD;

    $file = "$isoDate.md";
    file_put_contents($filePath, $content);
    echo "Done";
} else {
    echo "Error";
}
?>
