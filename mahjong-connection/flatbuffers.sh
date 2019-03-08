#!/usr/bin/env bash
fbsPath="/Users/xh/git/maxpanda/protomaxpanda"
projectPath="/Users/xh/gojob/src/mahjong-connection"
targetPath="$projectPath/fbs"
#subdirList="fbs info"
subdirList="fbs"


# 删除目标文件下的所有目录
for dir in `ls $targetPath`
do
	if [ -d "$targetPath/$dir" ];then
		rm -rf $targetPath/$dir
	fi
done


echo "cd $fbsPath"
cd $fbsPath

for subdir in $subdirList
do
	# 编译文件
	for file in `ls $fbsPath/$subdir`
	do
		if [ -d "$fbsPath/$subdir/$file" ];then
			rm -rf $fbsPath/$subdir/$file
		elif [ -f $1 ];then
			flatc --go "$subdir/$file"
		fi
	done
done
echo "编译文件完成..."

# 拷贝到目标文件夹
for subdir in $subdirList
do
filelist=`ls $fbsPath/$subdir`
	for file in $filelist
	do
		if [ -d "$fbsPath/$subdir/$file" ];then
			cp -R $fbsPath/$subdir/$file $targetPath/$file
			rm -rf $fbsPath/$subdir/$file
		fi
	done
done

echo "拷贝文件完成..."

# 安装应用
cd $projectPath
for subdir in `ls $targetPath`
do
	echo  "go install ./fbs/$subdir"
	go install ./fbs/$subdir
done

echo "安装应用完成..."
